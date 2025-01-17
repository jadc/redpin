package misc

import (
	"database/sql"
	"errors"
	"fmt"
    "log"
    "net/http"
    "strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jadc/redpin/database"
)

var (
    ALREADY_PINNED = errors.New("Message is already pinned")

    // Hashset of valid message types
    VALID_MSG_TYPE = map[discordgo.MessageType]struct{}{
        discordgo.MessageTypeDefault: {},
        discordgo.MessageTypeReply: {},
        discordgo.MessageTypeChatInputCommand: {},
        discordgo.MessageTypeThreadStarterMessage: {},
    }
)

type PinRequest struct {
    guildID string
    message *discordgo.Message
    reference *PinRequest
}

// Hashset of messages currently being pinned
// Helps prevent rapid reactions from pinning a message twice
var pinning = make(map[string]struct{})

// CreatePinRequest creates a copy of the message, and all messages it references, in its current state
func CreatePinRequest(discord *discordgo.Session, guild_id string, message *discordgo.Message) (*PinRequest, error) {
    // Skip messages that cannot feasibly be pinned
    if _, ok := VALID_MSG_TYPE[message.Type]; !ok {
        return nil, fmt.Errorf("This type of message cannot be pinned")
    }

    // Skip messages currently being pinned
    if _, ok := pinning[message.ID]; ok {
        return nil, ALREADY_PINNED
    }
    pinning[message.ID] = struct{}{}

    // Retrieve current config
    db, err := database.Connect()
    if err != nil {
        return nil, fmt.Errorf("Failed to connect to database: %v", err)
    }
    c := db.GetConfig(guild_id)

    // Query database for if message is already pinned
    a, b, err := db.GetPin(guild_id, message.ID)

    // Only throw up error if it's an actual error (not just row not found)
    if err != nil && err != sql.ErrNoRows {
        return nil, fmt.Errorf("Failed to fetch pin id for message '%s': %v", message.ID, err)
    }

    // Abort if message is already pinned
    if len(a) != 0 || len(b) != 0 {
        return nil, ALREADY_PINNED
    }

    // Create pin request
    req := &PinRequest{ guildID: guild_id, message: message }

    if c.ReplyDepth > 0 && message.MessageReference != nil {
        // Iteratively create pin requests for any messages this one references
        curr, depth := req, 0
        for ref := message.MessageReference; ref != nil && depth < c.ReplyDepth; ref = req.reference.message.MessageReference {
            // Fetch message that is being referenced
            ref_msg, err := discord.ChannelMessage(ref.ChannelID, ref.MessageID)
            if err != nil {
                log.Printf("Failed to fetch referenced message #%d: %v", depth, err)
                break
            }

            // Create pin request for said message
            curr.reference = &PinRequest{
                guildID: guild_id,
                message: ref_msg,
            }

            // Move pointer
            curr = curr.reference
            depth += 1
        }
    }

    log.Printf("Created new pin request for message '%s' in guild '%s'", message.ID, guild_id)
    return req, nil
}

// Execute on a PinRequest pins the message, forwarding it to the pin channel
// Returns the used pin channel ID and pin message's ID if successful
func (req *PinRequest) Execute(discord *discordgo.Session) (string, string, error) {
    // Create base webhook params
    params := &discordgo.WebhookParams{
        Username: "Unknown",
        AvatarURL: "",

        // Disable pinging
        AllowedMentions: &discordgo.MessageAllowedMentions{},
    }
    if a := req.message.Author; a != nil {
        if member, err := discord.GuildMember(req.guildID, a.ID); err == nil {
            params.Username = GetName(member)
            params.AvatarURL = member.AvatarURL("")
        }
    }

    // If the message being pinned is a reply, pin the referenced message first
    ref_pin_channel_id, ref_pin_msg_id := "", ""
    if req.reference != nil {
        ref_pin_channel_id, ref_pin_msg_id, _ = req.reference.Execute(discord)
    }

    // Get the current webhook
    webhook, err := GetWebhook(discord, req.guildID)
    if err != nil {
        return "", "", fmt.Errorf("Failed to retrieve webhook: %v", err)
    }

    // Send formatted link to pinned referenced message (if there is one)
    if ref_pin_channel_id != "" && ref_pin_msg_id != "" {
        params.Content = "-# ╰ Reply to " + GetMessageLink(req.guildID, ref_pin_channel_id, ref_pin_msg_id)
        _, err = discord.WebhookExecute(webhook.ID, webhook.Token, true, params)
        if err != nil {
            return "", "", fmt.Errorf("Failed to send reference header: %v", err)
        }
    }

    // Send the webhook copy to the pin channel
    pin_msg, err := req.cloneMessage(discord, webhook, params)
    if err != nil {
        return "", "", fmt.Errorf("Failed to clone pin message: %v", err)
    }

    // Send footer
    params.Content = "-# " + GetMessageLink(req.guildID, req.message.ChannelID, req.message.ID) + " " + req.message.Author.Mention()
    _, err = discord.WebhookExecute(webhook.ID, webhook.Token, true, params)
    if err != nil {
        return "", "", fmt.Errorf("Failed to send pin footer: %v", err)
    }

    // Copy reactions from original message if possible
    for _, r := range req.message.Reactions {
        discord.MessageReactionAdd(pin_msg.ChannelID, pin_msg.ID, r.Emoji.APIName())
    }

    // Add pin message to database
    db, err := database.Connect()
    if err != nil {
        return "", "", fmt.Errorf("Failed to connect to database: %v", err)
    }
    c := db.GetConfig(req.guildID)

    err = db.AddPin(req.guildID, c.Channel, req.message.ID, pin_msg.ID)
    if err != nil {
        return "", "", fmt.Errorf("Failed to add pin to database: %v", err)
    }
    delete(pinning, req.message.ID)

    log.Printf("Pinned message '%s' in guild '%s'", req.message.ID, req.guildID)
    return pin_msg.ChannelID, pin_msg.ID, nil
}

// cloneMessage recreates the given message into the given webhook with the given base parameters
// Returns the message object that the webhook sent (not including header/footer/attachment messages)
func (req *PinRequest) cloneMessage(discord *discordgo.Session, webhook *discordgo.Webhook, base *discordgo.WebhookParams) (*discordgo.Message, error) {
    var pin_msg *discordgo.Message
    skip := false

    // Create copy of message as webhook parameters
    params := *base
    params.Content = req.message.Content
    params.Components = req.message.Components

    // Only copy rich embeds, not embeds from links (Discord will add them itself)
    for _, e := range req.message.Embeds {
        if e.Type == discordgo.EmbedTypeRich {
            params.Embeds = append(params.Embeds, e)
        }
    }

    // Append as many attachments to webhook that can fit
    var file_sets [][]*discordgo.File
    var link_sets [][]string
    i, j := 0, 0

    if len(req.message.Attachments) > 0 {
        // Get file upload size limit of guild
        size_limit, err := sizeLimit(discord, webhook.GuildID)
        if err != nil {
            return nil, err
        }

        // Split files based on this size limit
        file_sets, link_sets = splitAttachments(req.message.Attachments, size_limit)

        // Messages must either have content or files to be sent, otherwise Discord errors
        // First, attempt to attach first set of files (not links) to pin message
        if len(file_sets) > 0 && len(file_sets[0]) > 0 {
            params.Files = file_sets[0]
            i += 1
        } else {
            if params.Content == "" {
                if len(link_sets) > 0 && len(link_sets[0]) > 0 {
                    // If message content is empty, try to add first link set
                    params.Content = strings.Join(link_sets[0], "\n")
                    j += 1
                } else {
                    // If there are no link sets, skip the pin message
                    skip = true
                }
            }
        }
    }

    // Send the webhook copy to the pin channel
    if !skip {
        var err error
        pin_msg, err = discord.WebhookExecute(webhook.ID, webhook.Token, true, &params)
        if err != nil {
            return nil, err
        }
    }

    // Send any attachment messages afterwards
    att := *base
    for {
        // Pop from files
        if i < len(file_sets) {
            att.Files = file_sets[i]
            i++
        } else {
            att.Files = nil
        }

        // Pop from links
        if j < len(link_sets) {
            att.Content = strings.Join(link_sets[j], "\n")
            j++
        } else {
            att.Content = ""
        }

        // Send attachment message
        if att.Files != nil || att.Content != "" {
            att_msg, err := discord.WebhookExecute(webhook.ID, webhook.Token, true, &att)
            if err != nil {
                return nil, err
            }

            // If pin message was skipped, set pin message to first attachment message
            if skip {
                pin_msg = att_msg
                skip = false
            }
        } else {
            break
        }
    }

    return pin_msg, nil
}

// splitAttachments splits a list of attachments into list of lists of attachments, each sublist under the size limit
func splitAttachments(attachments []*discordgo.MessageAttachment, size_limit int) ([][]*discordgo.File, [][]string) {
    var file_sets [][]*discordgo.File
    var link_sets [][]string

    files := make([]*discordgo.File, 0, MAX_FILES)
    links := make([]string, 0, MAX_LINKS)
    size := 0

    for _, a := range attachments {
        // Split links if getting too big
        if len(links) >= MAX_LINKS {
            link_sets = append(link_sets, links)
            links = make([]string, 0, MAX_LINKS)
        }

        if a.Size > 0 && a.Size < size_limit {
            // Download attachment
            data, err := http.DefaultClient.Get(a.URL)
            if err != nil {
                data, err = http.DefaultClient.Get(a.ProxyURL)
                if err != nil {
                    // Append link instead if downloading attachment fails
                    links = append(links, a.URL)
                    continue
                }
            }

            // Split files if getting too big
            if len(files) >= MAX_FILES || size + a.Size >= size_limit {
                file_sets = append(file_sets, files)
                files = make([]*discordgo.File, 0, MAX_FILES)
            }

            // Create file with attachment data
            file := &discordgo.File{
                Name: a.Filename,
                ContentType: a.ContentType,
                Reader: data.Body,
            }
            files = append(files, file)
            size += a.Size
        } else {
            // If an attachment is too big to fit in even one message, just append link to it
            links = append(links, a.URL)
        }
    }

    // Append any remaining files/links
    file_sets = append(file_sets, files)
    link_sets = append(link_sets, links)

    return file_sets, link_sets
}

// sizeLimit returns the maximum size (in bytes) of a message that can be sent in a guild
func sizeLimit(discord *discordgo.Session, guild_id string) (int, error) {
    // Get guild object
    guild, err := discord.Guild(guild_id)
    if err != nil {
        return 0, fmt.Errorf("Failed to retrieve guild '%s': %v", guild_id, err)
    }

    var mb int
    switch guild.PremiumTier {
        case 3:
            mb = 100
        case 2:
            mb = 50
        default:
            mb = 25
    }

    return mb * 1024 * 1024, nil
}
