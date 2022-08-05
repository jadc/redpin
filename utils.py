import discord
from discord.ui import View, Button
import config

class Pin():
    def __init__(self, bot, message, reaction='📌'):
        self.bot = bot
        self.message = message
        self.reaction = reaction
        self.guild_config = config.guild(message.guild.id)
        self.pin_channel = self.bot.get_channel(self.guild_config['channel'])

    async def get_webhook(self):
        for hook in await self.pin_channel.webhooks():
            if hook.user.id == self.bot.user.id:
                return hook

        # if no webhook, create one
        print('Created a new webhook')
        return await self.pin_channel.create_webhook(name = 'redpin', reason = 'redpin functionality')

    async def clone_reactions(self, source, target):
        for reaction in source.reactions:
            try:
                await target.add_reaction(reaction)
            except discord.errors.HTTPException:
                print('Failed to clone a reaction (unknown emoji or lacking permissions)')
        print('Cloned all reactions')

    async def notify_of_pin(self, pinned_msg):
        timestamp = discord.utils.format_dt(self.message.created_at, style='R')
        await self.message.author.send(
                content = f'A message you created {timestamp} in **{self.message.guild.name}** was pinned!',
                view = View().add_item( Button(label="Check it out", url=pinned_msg.jump_url) )
                )
        print('Sent a DM to the author of the newly pinned message')

    async def broadcast(self):
        if self.guild_config['channel'] is None:
            print('Pin attempted in guild with no pin channel, ignored.')
            return

        # Mark message as pinned
        await self.message.add_reaction(self.reaction)
        print('Marked a message as pinned.')

        hook = await self.get_webhook()

        # Convert all attachments that can fit into files, and reupload them
        files = [ await x.to_file( use_cached=True, spoiler=x.is_spoiler() ) for x in self.message.attachments if x.size < hook.guild.filesize_limit ]

        # If attachment is greater than bot is allowed, append link instead
        attachments = [ x for x in self.message.attachments if x.size >= hook.guild.filesize_limit ]
        content_w_files = self.message.content

        for att in attachments:
            if att.is_spoiler():
                content_w_files += f'\n|| {att.url} ||'
            else:
                content_w_files += f'\n{att.url}'

        # Convert stickers to URLs, as webhooks cannot send stickers
        for sticker in self.message.stickers:
            content_w_files += f'\n{sticker.url}'

        pinned_msg = await hook.send(
            wait = True,
            content = content_w_files,
            #embeds = self.message.embeds,  # odd behavior
            username = self.message.author.display_name,
            avatar_url = self.message.author.display_avatar.url,
            allowed_mentions = discord.AllowedMentions.none(),
            files = files,

            # jump to msg button
            view = View().add_item( Button(label="Jump", url=self.message.jump_url) )
        )

        await self.clone_reactions(self.message, pinned_msg)
        if self.guild_config['dm']:
            await self.notify_of_pin(pinned_msg)

        print(f'Pinned message {self.message.id} in guild {self.message.guild.id}')
        return True
