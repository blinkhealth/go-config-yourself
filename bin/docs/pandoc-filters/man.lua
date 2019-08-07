--[[
Pandoc filter to make nicer manpages that look well in github too
]]

local text = require('text')
function Header(elem)
    if elem.level == 1 then
        return pandoc.walk_block(elem, {
            Str = function(el)
                return pandoc.Str(text.upper(el.text))
            end
        })
    end
end

function Code(content_string)
    return pandoc.Strong(content_string)
end
