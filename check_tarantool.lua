#!/usr/bin/env tarantool

local function list_users()
    print("Listing all users:")
    local users = box.space._user:select()
    for _, user in pairs(users) do
        print(string.format("User: %s, ID: %s", user[2], user[1]))
    end
end

box.cfg{listen = 3301}
list_users()