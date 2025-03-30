box.cfg {
    listen = 3301,
    memtx_memory = 128 * 1024 * 1024, -- 128 MB
    wal_dir = '/var/lib/tarantool',
    memtx_dir = '/var/lib/tarantool',
    force_recovery = true
}

box.once('setup_users', function()
    -- Создаем пользователя admin
    box.schema.user.create('admin', {password = 'password', if_not_exists = true})
    box.schema.user.passwd('admin', 'password')
    -- Предоставляем привилегии
    box.schema.user.grant('admin', 'super', nil, nil, {if_not_exists = true})
    
    -- Также предоставляем гостевому пользователю необходимые права
    -- (если ваше приложение подключается как guest)
    box.schema.user.grant('guest', 'read,write,execute', 'universe', nil, {if_not_exists = true})
    
    print('Users setup completed')
end)

-- Create space for polls
local polls = box.schema.space.create('polls', {
    if_not_exists = true,
    format = {
        {name = 'id', type = 'string'},
        {name = 'title', type = 'string'},
        {name = 'options', type = 'array'},
        {name = 'creator_id', type = 'string'},
        {name = 'channel_id', type = 'string'},
        {name = 'created_at', type = 'datetime'},
        {name = 'finished_at', type = 'datetime', is_nullable = true},
        {name = 'is_finished', type = 'boolean'},
        {name = 'post_id', type = 'string'}
    }
})

-- Create indexes for polls
polls:create_index('primary', {
    type = 'hash',
    parts = {'id'},
    if_not_exists = true
})

polls:create_index('creator', {
    type = 'tree',
    parts = {'creator_id'},
    unique = false,
    if_not_exists = true
})

polls:create_index('channel', {
    type = 'tree',
    parts = {'channel_id'},
    unique = false,
    if_not_exists = true
})

-- Create space for votes
local votes = box.schema.space.create('votes', {
    if_not_exists = true,
    format = {
        {name = 'poll_id', type = 'string'},
        {name = 'user_id', type = 'string'},
        {name = 'option_idx', type = 'number'},
        {name = 'voted_at', type = 'datetime'}
    }
})

-- Create indexes for votes
votes:create_index('primary', {
    type = 'hash',
    parts = {'poll_id', 'user_id'},
    if_not_exists = true
})

votes:create_index('poll', {
    type = 'tree',
    parts = {'poll_id'},
    unique = false,
    if_not_exists = true
})

votes:create_index('user', {
    type = 'tree',
    parts = {'user_id'},
    unique = false,
    if_not_exists = true
})

votes:create_index('user_poll', {
    type = 'tree',
    parts = {'user_id', 'poll_id'},
    unique = true,
    if_not_exists = true
})

print('Tarantool initialized successfully')