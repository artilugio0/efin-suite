-- Allowed fields for validation
local allowed_fields = {
  id = true,
  timestamp = true,
  path = true,
  method = true,
  body = true,
  resp_status = true,
  resp_body = true,
  header = true,
  resp_header = true,
  raw = true,
  resp_raw = true,
}

-- Metatable for expressions (with .and_ and .or_ methods via __index)
local expr_mt = {
  __index = function(self, key)
    local mt = getmetatable(self)
    if key == 'and_' then
      return function(next_expr)
        return setmetatable({op = 'and', left = self, right = next_expr}, mt)
      end
    elseif key == 'or_' then
      return function(next_expr)
        return setmetatable({op = 'or', left = self, right = next_expr}, mt)
      end
    else
      error("Unknown method: " .. tostring(key))
    end
  end,

  __mt_id = 'expr_mt',
}

-- Metatable for fields (intercepts ops like .eq, .gt)
local field_mt = {
  __index = function(self, op_name)
    -- Map short op names to your preferred op strings (extend as needed)
    local op_map = {
      eq = 'eq',
      ne = 'ne',
      gt = 'gt',
      lt = 'lt',
      le = 'le',
      ge = 'ge',
      contains = 'contains'
      -- Add more, e.g., matches = 'matches' for regex
    }
    local op = op_map[op_name]
    if op then
      -- Return a function that builds the leaf expression when called
      return function(value)
        return setmetatable({field = self.field, op = op, value = value}, expr_mt)
      end
    else
      error("Unknown operation: " .. tostring(op_name))
    end
  end,

  __mt_id = 'field_mt',
}

-- The DSL entry point: q.field.op(value), plus q.not_(expr)
q = setmetatable({}, {
  __index = function(self, key)
    if key == 'not_' then
      -- Special handling for negation: q.not_(expr)
      return function(expr)
        return setmetatable({op = 'not', expr = expr}, expr_mt)
      end
    elseif key == 'header' or key == 'resp_header' then
      -- Special handling for header and resp_header: return a function to capture the header name
      return function(header_name)
        if type(header_name) ~= 'string' then
          error("Header name must be a string, got: " .. type(header_name))
        end
        return setmetatable({field = key .. ':' .. header_name}, field_mt)
      end
    else
      -- Validate field
      if not allowed_fields[key] then
        error("Invalid field: " .. tostring(key) .. ". Allowed fields are: path, method, body, resp_status, resp_body, header, resp_header")
      end
      -- Regular field access
      return setmetatable({field = key}, field_mt)
    end
  end
})

-- Optional: Function to pretty-print the AST for debugging
function dump_ast(t, indent)
  indent = indent or ''
  if type(t) ~= 'table' then return tostring(t) end
  local s = indent .. '{\n'
  for k, v in pairs(t) do
    s = s .. indent .. '  ' .. k .. ' = '
    if type(v) == 'table' then
      s = s .. dump_ast(v, indent .. '  ') .. ',\n'
    else
      s = s .. tostring(v) .. ',\n'
    end
  end
  return s .. indent .. '}'
end

-- Pretty-print a table with support for nested tables and arrays
function p(t, indent, visited)
  indent = indent or 0
  visited = visited or {}
  local indent_str = string.rep('  ', indent)

  -- Handle non-table types
  if type(t) ~= 'table' then
    if type(t) == 'string' then
      return string.format("%q", t) -- Quote strings properly
    elseif t == nil then
      return 'nil'
    elseif type(t) == 'boolean' or type(t) == 'number' then
      return tostring(t)
    else
      return '"[' .. type(t) .. ']"'
    end
  end

  -- Check for cyclic reference
  if visited[t] then
    return '"[cyclic reference]"'
  end
  visited[t] = true

  -- Determine if table is an array (sequential numeric keys starting from 1)
  local is_array = true
  local max_numeric_key = 0
  local count = 0
  for k, _ in pairs(t) do
    count = count + 1
    if type(k) ~= 'number' or k < 1 or k ~= math.floor(k) then
      is_array = false
    else
      max_numeric_key = math.max(max_numeric_key, k)
    end
  end
  is_array = is_array and count == max_numeric_key and max_numeric_key > 0

  local result = '{'
  local inner_indent = string.rep('  ', indent + 1)
  local items = {}

  if is_array then
    -- Handle array-like table, skipping nil values
    for i = 1, max_numeric_key do
      local v = t[i]
      if v ~= nil then
        if i == #items + 1 then
          -- Sequential key, omit index
          table.insert(items, inner_indent .. p(v, indent + 1, visited))
        else
          -- Sparse key, use explicit index
          table.insert(items, inner_indent .. '[' .. i .. '] = ' .. p(v, indent + 1, visited))
        end
      end
    end
  else
    -- Handle dictionary-like table
    local keys = {}
    for k in pairs(t) do table.insert(keys, k) end
    table.sort(keys, function(a, b)
      if type(a) == type(b) then return tostring(a) < tostring(b) end
      return type(a) < type(b)
    end)
    for _, k in ipairs(keys) do
      local v = t[k]
      local key_str
      if type(k) == 'string' then
        key_str = string.format("[%q]", k) -- Always quote string keys
      else
        key_str = '[' .. p(k, indent + 1, visited) .. ']'
      end
      table.insert(items, inner_indent .. key_str .. ' = ' .. p(v, indent + 1, visited))
    end
  end

  if #items > 0 then
    result = result .. '\n' .. table.concat(items, ',\n') .. '\n' .. indent_str
  end
  return result .. '}'
end

-- Deep copy a table, handling nested tables and cycles
-- TODO: remove this as it has been added into common definitions
function deep_copy(t, visited)
  -- Initialize visited table to track cyclic references
  visited = visited or {}

  -- Handle non-table types
  if type(t) ~= 'table' then
    return t
  end

  -- Check for cyclic reference
  if visited[t] then
    return nil -- Return nil for cycles (or could error/return placeholder)
  end
  visited[t] = true

  -- Create new table for the copy
  local copy = {}

  -- Copy all key-value pairs
  for k, v in pairs(t) do
    -- Recursively copy keys and values that are tables
    copy[deep_copy(k, visited)] = deep_copy(v, visited)
  end

  -- Copy the metatable, if any
  local mt = getmetatable(t)
  if mt then
    setmetatable(copy, deep_copy(mt, visited))
  end

  return copy
end
