import os, json

class Config():
    def __init__(self):
        self.config = {}
        self.config_name = 'config.json'
        self.load()
        print('Config initialized')

    def save(self):
        with open(self.config_name, 'w') as f:
            json.dump(self.config, f, ensure_ascii=False)

    def load(self):
        if os.path.isfile(self.config_name):
            with open(self.config_name, 'r') as f:
                self.config = json.load(f)
        else:
            self.save()
            self.load()

    def get(self, guild_id):
        default_guild = {'channel': None, 'count': 3, 'nsfw': False, 'selfpin': False, 'dm': True, 'filter': []}
        return self.config.setdefault( str(guild_id), default_guild )
