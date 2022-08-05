import discord
from discord.ext import commands
from discord.ui import View, Button
import asyncio

class Events(commands.Cog):
    def __init__(self, bot: commands.Bot, config):
        self.bot = bot
        self.config = config
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
        cfg = self.config.get(payload.guild_id)

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

            await msg.add_reaction(reaction)  # Marked as pinned
            print('Marked a message as pinned.')

            await self.pin_message(msg, self.bot.get_channel(cfg['channel'])) 
 
    # Filtering process
    def is_emoji_allowed(self, reaction):
        allow = self.config.get(reaction.message.guild.id)['filter']
        return len(allow) <= 0 or str(reaction) in allow

    async def get_real_count(self, reaction):
        if self.config.get(reaction.message.guild.id)['selfpin']:
            return reaction.count
        else:
            reactions_by_author = [user async for user in reaction.users() if user.id == reaction.message.author.id]
            return reaction.count - len(reactions_by_author)

    # Pinning process
    async def get_webhook(self, pin_channel):
        for hook in await pin_channel.webhooks():
            if hook.user.id == self.bot.user.id:
                return hook

        # if no webhook, create one
        print('Created a new webhook')
        return await pin_channel.create_webhook(name = 'redpin', reason = 'redpin functionality')

    async def pin_message(self, message, pin_channel):
        hook = await self.get_webhook(pin_channel)

        # Convert all attachments that can fit into files, and reupload them
        files = [ await x.to_file( use_cached=True, spoiler=x.is_spoiler() ) for x in message.attachments if x.size < hook.guild.filesize_limit ]

        # If attachment is greater than bot is allowed, append link instead
        attachments = [ x for x in message.attachments if x.size >= hook.guild.filesize_limit ]
        content_w_files = message.content

        for att in attachments:
            if att.is_spoiler():
                content_w_files += f'\n|| {att.url} ||'
            else:
                content_w_files += f'\n{att.url}'

        # Convert stickers to URLs, as webhooks cannot send stickers
        for sticker in message.stickers:
            content_w_files += f'\n{sticker.url}'

        print(f'Pinned message {message.id} in guild {message.guild.id}')
        pinned_msg = await hook.send(
            wait = True,
            content = content_w_files,
            #embeds = message.embeds,  # odd behavior
            username = message.author.display_name,
            avatar_url = message.author.display_avatar.url,
            allowed_mentions = discord.AllowedMentions.none(),
            files = files,

            # jump to msg button
            view = View().add_item( Button(label="Jump", url=message.jump_url) )
        )

        await self.clone_reactions(message, pinned_msg)
        if self.config.get(message.guild.id)['dm']:
            await self.notify_of_pin(message, pinned_msg)

    async def clone_reactions(self, source, target):
        for reaction in source.reactions:
            try:
                await target.add_reaction(reaction)
            except discord.errors.HTTPException:
                print('Failed to clone a reaction (unknown emoji or lacking permissions)')
        print('Cloned all reactions')

    async def notify_of_pin(self, message, pinned_msg):
        timestamp = discord.utils.format_dt(message.created_at, style='R')
        await message.author.send(
                content = f'A message you created {timestamp} in **{message.guild.name}** was pinned!',
                view = View().add_item( Button(label="Check it out", url=pinned_msg.jump_url) )
                )
        print('Sent a DM to the author of the newly pinned message')
