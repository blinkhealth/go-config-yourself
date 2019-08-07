--[[
Pandoc filter to make nicer manpages that look well in github too
]]

local text = require('text')
function Link(elem)
    if elem.target:match(".md$") then
        elem.target = elem.target:gsub(".md$", ".html"):gsub("README", "index")
        return elem
    end
end
