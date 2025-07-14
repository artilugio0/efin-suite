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
