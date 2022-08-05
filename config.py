import os, json
file_name = 'config.json'

def save():
    with open(file_name, 'w') as f:
        try:
            json.dump(config, f, ensure_ascii=False)
            print('Saved current configuration to file')
        except NameError:
            json.dump({}, f, ensure_ascii=False)
            print('Created a new configuration file')

def load():
    '''
    Usage:
        import config
        config = config.load()
    '''
    print('Config requested to be loaded')
    if os.path.isfile(file_name):
        print('Config file found, attempting to read')
        with open(file_name, 'r') as f:
            print('Config read')
            return json.load(f)
    else:
        print('No config file found')
        save()
        return load()

def guild(guild_id):
    default_guild = {'channel': None, 'count': 3, 'nsfw': False, 'selfpin': False, 'dm': True, 'filter': []}
    return config.setdefault( str(guild_id), default_guild )

config = load()
print('Per-guild configs initialized')
