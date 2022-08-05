import discord
from discord.ext import commands

from commands import Commands
from events import Events

class Bot(commands.Bot):
    def __init__(self):
        intents = discord.Intents.default()
        intents.reactions = True
        intents.message_content = True
        intents.members = True
        super().__init__(command_prefix = 'Born2PinForced2Prefix', intents=intents)

        # Set to True when developing
        self.synced = True

    async def setup_hook(self):
        await self.add_cog( Commands(self) )
        await self.add_cog( Events(self) )

        if not self.synced:
            await self.tree.sync()
            self.synced = True
            print('Synced')

    async def on_ready(self):
        await self.wait_until_ready()
        print('Ready')

if __name__ == "__main__":
    with open('token', 'r') as f:
        Bot().run( f.readline() )
