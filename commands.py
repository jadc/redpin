import discord
from discord import app_commands
from discord.ext import commands
from discord.ui import View, Button
import config
from utils import Pin

@app_commands.default_permissions(administrator=True)
class Commands(commands.GroupCog, name='redpin'):
    def __init__(self, bot: commands.Bot) -> None:
        super().__init__()
        self.bot = bot

        # Context menu specific hackery
        self.ctx_menu = app_commands.ContextMenu(name='Force Pin Message', callback=self.pin)
        self.bot.tree.add_command(self.ctx_menu)

        print('Commands initialized')

    # CONTEXT MENU
    async def cog_unload(self) -> None:
        self.bot.tree.remove_command(self.ctx_menu.name, type=self.ctx_menu.type)
        
    async def pin(self, interaction: discord.Interaction, message: discord.Message):
        await Pin(self.bot, message).broadcast()
        await interaction.response.send_message('Pinned a message.')

    # COMMANDS
    @app_commands.command(name = 'channel', description = 'Set which channel to send pins to.')
    async def channel(self, interaction: discord.Interaction, channel: discord.TextChannel):
        # if channel was set before
        if config.guild(interaction.guild_id)['channel'] is not None:
            # remove webhook from old channel
            old_channel = interaction.guild.get_channel( config.guild(interaction.guild_id)['channel'] )
            if old_channel is not None:
                for hook in await old_channel.webhooks():
                    if hook.user.id == self.bot.user.id:
                        await hook.delete(reason = 'Pin channel changed, webhook automatically removed')

        # update config
        config.guild(interaction.guild_id)['channel'] = channel.id
        config.save()

        # response
        await interaction.response.send_message(f'Pins will now be sent in <#{channel.id}>.', ephemeral = True)

    @app_commands.command(name = 'count', description = 'Set the number of reactions to pin a message.')
    async def count(self, interaction: discord.Interaction, count: int):

        # No negative or zero reaction count
        if count < 1:
            count = 1

        # update config
        config.guild(interaction.guild_id)['count'] = count
        config.save()

        plural = 's'
        if count == 1: plural = ''

        # response
        await interaction.response.send_message(f'Pins will now require **{count} reaction{plural}**.', ephemeral = True)

    @app_commands.command(name = 'nsfw', description = 'Toggle whether messages from NSFW channels can be pinned.')
    async def nsfw(self, interaction: discord.Interaction):
        # update config
        config.guild(interaction.guild_id)['nsfw'] = not config.guild(interaction.guild_id)['nsfw']
        config.save()

        # response
        if config.guild(interaction.guild_id)['nsfw']:
            await interaction.response.send_message(f'Messages from NSFW channels can now be pinned.', ephemeral = True)
        else:
            await interaction.response.send_message(f'Messages from NSFW channels can no longer be pinned.', ephemeral = True)

    @app_commands.command(name = 'selfpin', description = 'Toggle whether messages can be pinned by their author.')
    async def selfpin(self, interaction: discord.Interaction):
        # update config
        config.guild(interaction.guild_id)['selfpin'] = not config.guild(interaction.guild_id)['selfpin']
        config.save()

        # response
        if config.guild(interaction.guild_id)['selfpin']:
            await interaction.response.send_message(f'Messages can now be pinned by their author.', ephemeral = True)
        else:
            await interaction.response.send_message(f'Messages can no longer be pinned by their author.', ephemeral = True)

    @app_commands.command(name = 'dm', description = 'Toggle whether pinning a message notifies their author.')
    async def dm(self, interaction: discord.Interaction):
        # update config
        config.guild(interaction.guild_id)['dm'] = not self.config.guild(interaction.guild_id)['dm']
        config.save()

        # response
        if config.guild(interaction.guild_id)['dm']:
            await interaction.response.send_message(f'Pinning a message now notifies their author.', ephemeral = True)
        else:
            await interaction.response.send_message(f'Pinning a message no longer notifies their author.', ephemeral = True)

    # EMOJI COMMAND
    @app_commands.command(name = 'filter', description = 'Customize which emojis can pin messages. Run this command in a private channel!')
    async def filter(self, interaction: discord.Interaction):
        view = EmojiPrompt()

        await interaction.response.send_message('**Customize which emojis can pin messages.**\nReact to this message with the emojis you want to be able to pin messages with.\n*Submit with no reactions to allow any emoji to pin messages.*', view=view)
        await view.wait()

        if view.value:
            prompt = await interaction.original_message()
            config.guild(interaction.guild_id)['filter'] = [str(x) for x in prompt.reactions]
            config.save()
            await interaction.delete_original_message()

class EmojiPrompt(discord.ui.View):
    def __init__(self):
        super().__init__()
        self.value = None

    @discord.ui.button(label='Confirm', style=discord.ButtonStyle.green)
    async def confirm(self, interaction: discord.Interaction, button: discord.ui.Button):
        await interaction.response.send_message('Saved changes.', ephemeral=True)
        self.value = True
        self.stop()

    @discord.ui.button(label='Cancel', style=discord.ButtonStyle.grey)
    async def cancel(self, interaction: discord.Interaction, button: discord.ui.Button):
        await interaction.response.send_message('Cancelled. Nothing was changed.', ephemeral=True)
        self.value = False
        self.stop()
