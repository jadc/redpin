import discord
from discord.ext import commands
from discord.ui import View, Button
import asyncio
import config
from utils import Pin

class Events(commands.Cog):
    def __init__(self, bot: commands.Bot):
        self.bot = bot
        self.lock = asyncio.Lock()
        print('Events initialized')

    @commands.Cog.listener()
    async def on_raw_reaction_add(self, payload):
        if payload.guild_id is None:
            print('Reaction in DMs, ignored.')
            return

        if payload.member.bot:
            print('Reaction from bot, ignored.')
            return

        async with self.lock:
            await self.handle_reaction(payload)

    async def handle_reaction(self, payload):
        cfg = config.guild(payload.guild_id)

        if cfg['channel'] is None:
            print('Reaction from guild with no pin channel, ignored.')
            return

        if payload.channel_id == cfg['channel']:
            print('Reaction in pin channel, ignored.')
            return

        channel = self.bot.get_channel(payload.channel_id)

        if not cfg['nsfw'] and channel.is_nsfw():
            print('Reaction in NSFW channel, ignored.')
            return
 
        msg = await channel.fetch_message(payload.message_id)

        if any(x.me for x in msg.reactions):
            print('Message already pinned, ignored.')
            return

        # Filters
        pin_reactions = [x for x in msg.reactions if await self.get_real_count(x) >= cfg['count'] and self.is_emoji_allowed(x)]

        if pin_reactions:
            # Does not matter which reaction is selected
            reaction = pin_reactions[0]

            await Pin(self.bot, msg, reaction).broadcast()
 
    # Filtering process
    def is_emoji_allowed(self, reaction):
        allow = config.guild(reaction.message.guild.id)['filter']
        return len(allow) <= 0 or str(reaction) in allow

    async def get_real_count(self, reaction):
        if config.guild(reaction.message.guild.id)['selfpin']:
            return reaction.count
        else:
            reactions_by_author = [user async for user in reaction.users() if user.id == reaction.message.author.id]
            return reaction.count - len(reactions_by_author)
