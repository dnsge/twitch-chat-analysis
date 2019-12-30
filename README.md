# Twitch Chat Analysis
Program that connects to a twitch.tv IRC chat and collects information on emote
usage vs. months subscribed. Supports collection of Twitch emotes, BTTV emotes,
and FFZ emotes. 

Program allows for collection with saving for later manual analysis or an
interactive collection and analysis mode.

### Usage Instructions
Build from source
```shell script
go build twitch-chat-analysis
```

CLI Arguments
```
 -c            # Channel Name on Twitch

 -i            # Input file (for analyze)

 -o            # Output file (for collect)

 -t            # Time to run collection, in seconds

 -print true   # Print emotes as they are collected

 -unique true  # Only count each emote once per message,
               # i.e. 4 LULW's are counted as 1
```

Example
```shell script
# Collect and analyze later
./twitch-chat-analysis -c forsen -o forsen_data.json -t 600 collect 
./twitch-chat-analysis -c forsen -i forsen_data.json analyze

# Interactive collection and analysis
./twitch-chat-analysis -unique=true interactive
```