do
    local s = "hello"
    print(string.byte(s))
    --> =104

    print(string.byte(s, -1))
    --> =111

    print(string.byte(s, 2, 4))
    --> =101	108	108

    print(string.byte(s, -5, 2))
    --> =104	101
    
    print(string.byte(s, -2, -1))
    --> =108	111

    -- Byte can also be called as a method on strings
    print(s:byte(3))
    --> =108
end
