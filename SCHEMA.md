Messages (create table per guild)
---
Original Message ID | Pin Message Copy Message ID

Stats
---
User ID | Guild ID | Emoji used to pin one of their messages | Emoji Name (fallback)

Settings
---
Guild ID | serialized config (jsonb)

json config:

```json
{
    "channel": <pin channel>,
    "count": <reactions needed to pin>,
    "nsfw": <pin nsfw msgs>,
    "selfpin": <allow self pin>,
    "allowlist": [<list of emojis that pin, if empty, any pin>]
}
```
