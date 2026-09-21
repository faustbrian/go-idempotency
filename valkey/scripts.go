package valkey

import (
	"strconv"
	"strings"

	"github.com/faustbrian/go-idempotency"
	valkeygo "github.com/valkey-io/valkey-go"
)

var nativeScripts = map[operation]*valkeygo.Lua{
	operationAcquire:   valkeygo.NewLuaScript(acquireScript),
	operationInspect:   valkeygo.NewLuaScriptReadOnly(inspectScript),
	operationHeartbeat: valkeygo.NewLuaScript(heartbeatScript),
	operationComplete:  valkeygo.NewLuaScript(completeScript),
	operationFail:      valkeygo.NewLuaScript(failScript),
	operationRelease:   valkeygo.NewLuaScript(releaseScript),
	operationExpire:    valkeygo.NewLuaScript(expireScript),
}

const recordValidationPrelude = `
local record_fields = {
    'schema', 'namespace', 'tenant', 'operation', 'caller', 'key_value',
    'fingerprint_version', 'fingerprint_sum', 'state', 'owner_token',
    'fencing_token', 'lease_expires_at_ms', 'heartbeat_at_ms', 'attempt',
    'created_at_ms', 'updated_at_ms', 'completed_at_ms', 'failed_at_ms',
    'abandoned_at_ms', 'expired_at_ms', 'result', 'metadata'
}

local field_limits = {
    schema = 1,
    namespace = {{max_key_part}}, tenant = {{max_key_part}},
    operation = {{max_key_part}}, caller = {{max_key_part}},
    key_value = {{max_key_part}},
    fingerprint_version = {{max_fingerprint_version}}, fingerprint_sum = 64,
    state = 9, owner_token = {{max_owner_token}}, fencing_token = 20,
    lease_expires_at_ms = 19, heartbeat_at_ms = 19, attempt = 20,
    created_at_ms = 19, updated_at_ms = 19, completed_at_ms = 19,
    failed_at_ms = 19, abandoned_at_ms = 19, expired_at_ms = 19,
    result = {{max_result}}, metadata = {{max_encoded_metadata}}
}

local function semantic_error(reason)
    return {'error', reason}
end

local function bounded_uint(value, maximum, positive)
    if not string.match(value, '^%d+$') then return false end
    local normalized = string.gsub(value, '^0+', '')
    if normalized == '' then return not positive end
    if #normalized < #maximum then return true end
    if #normalized > #maximum then return false end
    return normalized <= maximum
end

local function valid_metadata(encoded)
    if string.match(encoded, '^%s*null%s*$') then return true end
    if not string.match(encoded, '^%s*{') or not string.match(encoded, '}%s*$') then
        return false
    end
    local ok, decoded = pcall(cjson.decode, encoded)
    if not ok or type(decoded) ~= 'table' then return false end
    local count = 0
    for key, value in pairs(decoded) do
        if type(key) ~= 'string' or type(value) ~= 'string' then return false end
        count = count + 1
        if count > {{max_metadata_entries}} or
           #key > {{max_metadata_key}} or #value > {{max_metadata_value}} then
            return false
        end
    end
    return true
end

local function validate_record()
    local count_ok, count = pcall(redis.call, 'HLEN', KEYS[1])
    if not count_ok or count ~= #record_fields then
        return semantic_error('invalid_payload')
    end
    for _, field in ipairs(record_fields) do
        if redis.call('HEXISTS', KEYS[1], field) ~= 1 then
            return semantic_error('invalid_payload')
        end
        local length_ok, length = pcall(redis.call, 'HSTRLEN', KEYS[1], field)
        if not length_ok then return semantic_error('invalid_payload') end
        if length > field_limits[field] then
            return semantic_error('limit_exceeded')
        end
    end

    local values = redis.call('HMGET', KEYS[1], unpack(record_fields))
    local fields = {}
    for index, field in ipairs(record_fields) do fields[field] = values[index] end

    if fields.schema ~= '1' or
       fields.namespace == '' or fields.namespace ~= ARGV[1] or
       fields.tenant == '' or fields.tenant ~= ARGV[2] or
       fields.operation == '' or fields.operation ~= ARGV[3] or
       fields.caller == '' or fields.caller ~= ARGV[4] or
       fields.key_value == '' or fields.key_value ~= ARGV[5] or
       fields.fingerprint_version == '' or
       not string.match(fields.fingerprint_sum, '^[0-9a-fA-F]+$') or
       #fields.fingerprint_sum ~= 64 or fields.owner_token == '' then
        return semantic_error('invalid_payload')
    end
    if fields.state ~= 'acquired' and fields.state ~= 'running' and
       fields.state ~= 'completed' and fields.state ~= 'failed' and
       fields.state ~= 'expired' and fields.state ~= 'abandoned' then
        return semantic_error('invalid_payload')
    end
    if not bounded_uint(fields.fencing_token, '18446744073709551615', true) or
       not bounded_uint(fields.attempt, '18446744073709551615', true) then
        return semantic_error('invalid_payload')
    end
    local timestamps = {
        'lease_expires_at_ms', 'heartbeat_at_ms', 'created_at_ms', 'updated_at_ms',
        'completed_at_ms', 'failed_at_ms', 'abandoned_at_ms', 'expired_at_ms'
    }
    for _, field in ipairs(timestamps) do
        if not bounded_uint(fields[field], '9223372036854775807', false) then
            return semantic_error('invalid_payload')
        end
    end
    if not valid_metadata(fields.metadata) then
        return semantic_error('invalid_payload')
    end
    return nil
end

local function record_reply(outcome)
    local reply = {outcome}
    local values = redis.call('HMGET', KEYS[1], unpack(record_fields))
    for index, field in ipairs(record_fields) do
        table.insert(reply, field)
        table.insert(reply, values[index])
    end
    return reply
end
`

func recordScript(body string) string {
	return strings.NewReplacer(
		"{{max_key_part}}", strconv.Itoa(idempotency.MaxKeyPartBytes),
		"{{max_fingerprint_version}}", strconv.Itoa(idempotency.MaxFingerprintVersionBytes),
		"{{max_owner_token}}", strconv.Itoa(idempotency.MaxOwnerTokenBytes),
		"{{max_result}}", strconv.Itoa(idempotency.MaxResultBytes),
		"{{max_encoded_metadata}}", strconv.Itoa(maxEncodedMetadataBytes),
		"{{max_metadata_entries}}", strconv.Itoa(idempotency.MaxMetadataEntries),
		"{{max_metadata_key}}", strconv.Itoa(idempotency.MaxMetadataKeyBytes),
		"{{max_metadata_value}}", strconv.Itoa(idempotency.MaxMetadataValueBytes),
	).Replace(recordValidationPrelude + body)
}

var acquireScript = recordScript(`
local function now_ms()
    local current = redis.call('TIME')
    return current[1] * 1000 + math.floor(current[2] / 1000)
end
local now = now_ms()
local lease_ms = tonumber(ARGV[7])
local retention_ms = tonumber(ARGV[8])
if redis.call('EXISTS', KEYS[1]) == 0 then
    redis.call('HSET', KEYS[1],
        'schema', '1', 'namespace', ARGV[1], 'tenant', ARGV[2],
        'operation', ARGV[3], 'caller', ARGV[4], 'key_value', ARGV[5],
        'fingerprint_version', ARGV[9], 'fingerprint_sum', ARGV[10],
        'state', 'acquired', 'owner_token', ARGV[6], 'fencing_token', '1',
        'lease_expires_at_ms', tostring(now + lease_ms),
        'heartbeat_at_ms', tostring(now), 'attempt', '1',
        'created_at_ms', tostring(now), 'updated_at_ms', tostring(now),
        'completed_at_ms', '0', 'failed_at_ms', '0',
        'abandoned_at_ms', '0', 'expired_at_ms', '0',
        'result', '', 'metadata', '{}')
    redis.call('PEXPIRE', KEYS[1], lease_ms + retention_ms)
    return record_reply('acquired')
end

local invalid = validate_record()
if invalid then return invalid end
if redis.call('HGET', KEYS[1], 'fingerprint_version') ~= ARGV[9] or
   redis.call('HGET', KEYS[1], 'fingerprint_sum') ~= ARGV[10] then
    return record_reply('conflict')
end
local state = redis.call('HGET', KEYS[1], 'state')
if state == 'completed' then return record_reply('replayed') end
if state == 'failed' then return record_reply('terminal_failure') end
if (state == 'acquired' or state == 'running') and
   tonumber(redis.call('HGET', KEYS[1], 'lease_expires_at_ms')) > now then
    return record_reply('in_progress')
end
local fence = redis.call('HGET', KEYS[1], 'fencing_token')
local attempt = redis.call('HGET', KEYS[1], 'attempt')
if not bounded_uint(fence, '9223372036854775806', true) or
   not bounded_uint(attempt, '9223372036854775806', true) then
    return semantic_error('limit_exceeded')
end
local outcome = 'acquired'
if state == 'acquired' or state == 'running' then outcome = 'stale_owner_takeover' end
redis.call('HINCRBY', KEYS[1], 'fencing_token', 1)
redis.call('HINCRBY', KEYS[1], 'attempt', 1)
redis.call('HSET', KEYS[1],
    'state', 'acquired', 'owner_token', ARGV[6],
    'lease_expires_at_ms', tostring(now + lease_ms),
    'heartbeat_at_ms', tostring(now), 'updated_at_ms', tostring(now),
    'completed_at_ms', '0', 'failed_at_ms', '0',
    'abandoned_at_ms', '0', 'expired_at_ms', '0',
    'result', '', 'metadata', '{}')
redis.call('PEXPIRE', KEYS[1], lease_ms + retention_ms)
return record_reply(outcome)
`)

var inspectScript = recordScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then return semantic_error('not_found') end
local invalid = validate_record()
if invalid then return invalid end
return record_reply('ok')
`)

var heartbeatScript = recordScript(`
local function now_ms()
    local current = redis.call('TIME')
    return current[1] * 1000 + math.floor(current[2] / 1000)
end
if redis.call('EXISTS', KEYS[1]) == 0 then return semantic_error('not_found') end
local invalid = validate_record()
if invalid then return invalid end
if redis.call('HGET', KEYS[1], 'owner_token') ~= ARGV[6] or
   redis.call('HGET', KEYS[1], 'fencing_token') ~= ARGV[7] then
    return semantic_error('stale_owner')
end
local state = redis.call('HGET', KEYS[1], 'state')
if state ~= 'acquired' and state ~= 'running' then return semantic_error('invalid_transition') end
local now = now_ms()
if now >= tonumber(redis.call('HGET', KEYS[1], 'lease_expires_at_ms')) then
    return semantic_error('lease_expired')
end
local lease_ms = tonumber(ARGV[8])
redis.call('HSET', KEYS[1], 'state', 'running',
    'heartbeat_at_ms', tostring(now), 'lease_expires_at_ms', tostring(now + lease_ms),
    'updated_at_ms', tostring(now))
redis.call('PEXPIRE', KEYS[1], lease_ms + tonumber(ARGV[9]))
return record_reply('ok')
`)

var completeScript = recordScript(`
local function now_ms()
    local current = redis.call('TIME')
    return current[1] * 1000 + math.floor(current[2] / 1000)
end
if redis.call('EXISTS', KEYS[1]) == 0 then return semantic_error('not_found') end
local invalid = validate_record()
if invalid then return invalid end
if redis.call('HGET', KEYS[1], 'owner_token') ~= ARGV[6] or
   redis.call('HGET', KEYS[1], 'fencing_token') ~= ARGV[7] then
    return semantic_error('stale_owner')
end
local state = redis.call('HGET', KEYS[1], 'state')
if state ~= 'acquired' and state ~= 'running' then return semantic_error('invalid_transition') end
local now = now_ms()
if now >= tonumber(redis.call('HGET', KEYS[1], 'lease_expires_at_ms')) then
    return semantic_error('lease_expired')
end
redis.call('HSET', KEYS[1], 'state', 'completed', 'result', ARGV[8],
    'metadata', ARGV[9], 'completed_at_ms', tostring(now), 'updated_at_ms', tostring(now))
redis.call('PEXPIRE', KEYS[1], tonumber(ARGV[10]))
return record_reply('ok')
`)

var failScript = recordScript(`
local function now_ms()
    local current = redis.call('TIME')
    return current[1] * 1000 + math.floor(current[2] / 1000)
end
if redis.call('EXISTS', KEYS[1]) == 0 then return semantic_error('not_found') end
local invalid = validate_record()
if invalid then return invalid end
if redis.call('HGET', KEYS[1], 'owner_token') ~= ARGV[6] or
   redis.call('HGET', KEYS[1], 'fencing_token') ~= ARGV[7] then
    return semantic_error('stale_owner')
end
local state = redis.call('HGET', KEYS[1], 'state')
if state ~= 'acquired' and state ~= 'running' then return semantic_error('invalid_transition') end
local now = now_ms()
if now >= tonumber(redis.call('HGET', KEYS[1], 'lease_expires_at_ms')) then
    return semantic_error('lease_expired')
end
redis.call('HSET', KEYS[1], 'state', 'failed', 'result', ARGV[8],
    'metadata', ARGV[9], 'failed_at_ms', tostring(now), 'updated_at_ms', tostring(now))
redis.call('PEXPIRE', KEYS[1], tonumber(ARGV[10]))
return record_reply('ok')
`)

var releaseScript = recordScript(`
local function now_ms()
    local current = redis.call('TIME')
    return current[1] * 1000 + math.floor(current[2] / 1000)
end
if redis.call('EXISTS', KEYS[1]) == 0 then return semantic_error('not_found') end
local invalid = validate_record()
if invalid then return invalid end
if redis.call('HGET', KEYS[1], 'owner_token') ~= ARGV[6] or
   redis.call('HGET', KEYS[1], 'fencing_token') ~= ARGV[7] then
    return semantic_error('stale_owner')
end
local state = redis.call('HGET', KEYS[1], 'state')
if state ~= 'acquired' and state ~= 'running' then return semantic_error('invalid_transition') end
local now = now_ms()
if now >= tonumber(redis.call('HGET', KEYS[1], 'lease_expires_at_ms')) then
    return semantic_error('lease_expired')
end
redis.call('HSET', KEYS[1], 'state', 'abandoned',
    'abandoned_at_ms', tostring(now), 'updated_at_ms', tostring(now))
redis.call('PEXPIRE', KEYS[1], tonumber(ARGV[8]))
return record_reply('ok')
`)

var expireScript = recordScript(`
local function now_ms()
    local current = redis.call('TIME')
    return current[1] * 1000 + math.floor(current[2] / 1000)
end
if redis.call('EXISTS', KEYS[1]) == 0 then return semantic_error('not_found') end
local invalid = validate_record()
if invalid then return invalid end
local state = redis.call('HGET', KEYS[1], 'state')
local now = now_ms()
if (state ~= 'acquired' and state ~= 'running') or
   now < tonumber(redis.call('HGET', KEYS[1], 'lease_expires_at_ms')) then
    return semantic_error('invalid_transition')
end
redis.call('HSET', KEYS[1], 'state', 'expired',
    'expired_at_ms', tostring(now), 'updated_at_ms', tostring(now))
redis.call('PEXPIRE', KEYS[1], tonumber(ARGV[6]))
return record_reply('ok')
`)
