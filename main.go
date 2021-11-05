package main 

import (
    "fmt"
    "github.com/bwmarrin/discordgo"
    "strings"
    "os"
    "strconv"
    goaway "github.com/TwiN/go-away"
    "time"
)

const mute string = "903591674016579614"
var ID string 
const prefix string = "#guard"

func main() {
    token := os.Getenv("TOKEN") 
    discord, err := discordgo.New("Bot " + token) 
    if err != nil {
        panic(err.Error())
    }
    user, err := discord.User("@me") 
    if err != nil {
        panic(err.Error())
    }
    ID = user.ID 
    discord.AddHandler(commands)
    discord.AddHandler(verification)    
    if err = discord.Open(); err != nil {
        panic(err.Error())
    }
    fmt.Println("The bot is running.")
    defer discord.Close()
    <- make(chan struct{})
}

func commands(session *discordgo.Session, message *discordgo.MessageCreate) {
    if message.Author.ID == ID {
        return 
    }
    if goaway.IsProfane(message.Content) {
        session.ChannelMessageDelete(message.ChannelID, message.ID)
        user, err := session.UserChannelCreate(message.Author.ID)
        if err != nil {
            panic(err.Error())
        }
        profane := goaway.IsProfane(message.Content)
        file, err := os.Open("guard.webp")
        embed := &discordgo.MessageEmbed{
            Title: ":warning: You have been warned.",
            Description:"Bad Words." ,
            Image: &discordgo.MessageEmbedImage{URL:"attachment://" + file.Name()},
        }
        if !profane {
            embed.Description = "Capitalization."
        }
        msg := &discordgo.MessageSend{
            Embed:embed,
            Files: []*discordgo.File{
                &discordgo.File{
                    Name:file.Name(),
                    Reader:file,
                },
            },
        }
        session.ChannelMessageSendComplex(user.ID, msg)
        file.Close()        
        return 
    } 
    instructions := strings.Fields(message.Content)
    if len(instructions) < 2 || instructions[0] != prefix {
        return 
    }
    if instructions[1] == "mute" {
        if len(instructions) == 2 || len(message.Mentions) != 1 {
            session.ChannelMessageSendReply(message.ChannelID, "That is the wrong command. The correct command is `" + prefix + " mute <member> <reason>`.", message.Reference())
            return
        }
        capable := Permissions(session, message.GuildID, message.Author.ID, discordgo.PermissionManageRoles)
        if capable {
            reason := "none"
            if len(instructions) > 3 {
                reason = ""
                for index := 3; index < len(instructions); index++ {
                    reason += instructions[index] + " "
                }
            }
            mention := message.Mentions[0]
            member, err := session.GuildMember(message.GuildID, mention.ID)
            if err != nil {
                panic(err.Error())
            }
            for _, role := range member.Roles {
                if role == mute {
                    session.ChannelMessageSendReply(message.ChannelID, "The mentioned user is already muted.", message.Reference())
                    return 
                }
            }
            err = session.GuildMemberRoleAdd(message.GuildID, mention.ID, mute)
            if err != nil {
                panic(err.Error())
            }
            embed := &discordgo.MessageEmbed{
                Title:mention.String() + " has been muted.",
                Description: "Reason: " + reason + "\nModerator: " + message.Author.Mention(),
            }
            session.ChannelMessageSendEmbed(message.ChannelID, embed)
        } else {
            session.ChannelMessageSendReply(message.ChannelID, "You need the `Manage Roles` permission to do that.", message.Reference())
        }
    } else if instructions[1] == "unmute" {
        if len(instructions) == 2 || len(message.Mentions) != 1 {
            session.ChannelMessageSendReply(message.ChannelID,"That is the wrong command. The correct command is `" + prefix + " unmute <member>`.", message.Reference())
            return 
        }
        capable := Permissions(session, message.GuildID, message.Author.ID, discordgo.PermissionManageRoles)
        if capable {
            mention := message.Mentions[0]
            member, err := session.GuildMember(message.GuildID, mention.ID)
            if err != nil {
                panic(err.Error())
            }
            muted := false 
            for _, role := range member.Roles {
                if role == mute {
                    muted = true 
                    break 
                }
            }
            if !muted {
                session.ChannelMessageSendReply(message.ChannelID, "The mentioned user is already unmuted.", message.Reference())
            } else {
                session.GuildMemberRoleRemove(message.GuildID, mention.ID, mute)
                embed := &discordgo.MessageEmbed{
                    Title:mention.String() + " has been unmuted.",
                    Description:"Moderator: " + message.Author.Mention(),
                }
                session.ChannelMessageSendEmbed(message.ChannelID, embed)
            }
        } else {
            session.ChannelMessageSendReply(message.ChannelID, "You need the 'Manage Roles` permission to do that.", message.Reference())
        }
    } else if instructions[1] == "kick" {
        if len(instructions) < 3 || len(message.Mentions) != 1 {
            session.ChannelMessageSendReply(message.ChannelID, "That is the wrong command. The correct command is `" + prefix + " kick <member> <reason>`.", message.Reference())
            return 
        }
        capable := Permissions(session, message.GuildID, message.Author.ID, discordgo.PermissionManageRoles) 
        if capable {
            var reason string = "none" 
            if len(instructions) > 3 {
                reason = ""
                for i := 3; i < len(instructions); i++ {
                    reason += instructions[i] + " "
                }
            }
            mention := message.Mentions[0]
            embed := &discordgo.MessageEmbed{
                Title :  mention.String() + " has been kicked from the server.",
                Description : "Reasons: " + strings.Title(reason) + "\nModerator: " + message.Author.Mention(),
            }
            session.ChannelMessageSendEmbed(message.ChannelID, embed)
            session.GuildMemberDelete(message.GuildID, mention.ID)
        } else {
            session.ChannelMessageSendReply(message.ChannelID, "You need the `Kick Members` permission to do that.", message.Reference())
        }
    } else if instructions[1] == "ban" {
        if len(instructions) < 3 || len(message.Mentions) != 1 {
            session.ChannelMessageSendReply(message.ChannelID, "That is the wrong command. The correct command is `" + prefix + " ban <member> <reason>`.", message.Reference())
            return 
        }
        capable := Permissions(session, message.GuildID, message.Author.ID, discordgo.PermissionBanMembers) 
        if capable {
            mention := message.Mentions[0]
            reason := "none"
            if len(instructions) > 3 {
                reason = ""
                for i := 3; i < len(instructions); i++ {
                    reason += instructions[i] + " "
                }
            }
            if reason == "none" {
                session.GuildBanCreate(message.GuildID, mention.ID, 7)
            } else {
                session.GuildBanCreateWithReason(message.GuildID, mention.ID,reason, 7)
            }
            embed := &discordgo.MessageEmbed{
                Title:mention.String() + " has been banned from the server.",
                Description:"Reason: " + reason + "\nModerator: " + message.Author.Mention(),
            }
            session.ChannelMessageSendEmbed(message.ChannelID, embed)
        } else {
            session.ChannelMessageSendReply(message.ChannelID, "You need the `Ban Members` permission to do that.", message.Reference())
        }
    } else if instructions[1] == "unban" {
        if len(instructions) == 2 || len(message.Mentions) != 0 {
            session.ChannelMessageSendReply(message.ChannelID,"That is the wrong command. The correct command is `" + prefix + " unmute <id>`.", message.Reference())
        }
        capable := Permissions(session, message.GuildID, message.Author.ID, discordgo.PermissionBanMembers)
        if capable {
            mention, _ := session.User(instructions[2])
            embed := &discordgo.MessageEmbed{
                Title:mention.String() + " has been unbanned from the server.",
                Description:"Moderator: " + message.Author.Mention(),
            }
            session.ChannelMessageSendEmbed(message.ChannelID, embed)
            session.GuildBanDelete(message.GuildID, mention.ID)
        } else {
            session.ChannelMessageSendReply(message.ChannelID, "You need the `Ban Members` permission to do that.", message.Reference())
        }

    } else if instructions[1] == "clear" {
        if len(instructions) > 3 {
            session.ChannelMessageSendReply(message.ChannelID, "That is the wrong command. The correct command is `" + prefix + " clear <no. of messages>`", message.Reference())
            return 
        }
        capable := Permissions(session, message.GuildID, message.Author.ID, discordgo.PermissionManageMessages)
        if capable {
            number := 100 
            if len(instructions) > 2 {
                number, _ = strconv.Atoi(instructions[2])
            }
            messages, err := session.ChannelMessages(message.ChannelID, number, "", "", "")
            if err != nil {
                panic(err.Error())
            }
            msgs := make([]string, len(messages))
            if len(msgs) < 1 {
                return 
            } 
            for index, chat := range messages {
                msgs[index] = chat.ID 
            }
            session.ChannelMessagesBulkDelete(message.ChannelID, msgs) 
            embed := &discordgo.MessageEmbed{
                Description:"```Deleted " + strconv.Itoa(len(msgs)) + " messages from the channel.```",
            }
            session.ChannelMessageSendEmbed(message.ChannelID, embed)
            time.Sleep(time.Millisecond * 500) 
            msg, _ := session.ChannelMessages(message.ChannelID, 1, "", "", "")
            session.ChannelMessageDelete(message.ChannelID, msg[0].ID)
        } else {
            session.ChannelMessageSendReply(message.ChannelID, "You need the `Manage Messages` permission to do that.", message.Reference())
        }
    } else if instructions[1] == "bot" {
        Title := "The Pub's Gaurd."
        Description:= "_**The Pub's Guard is a moderator bot.\n That is specifically designed and\n developed with Pablo's Pub in mind. \nThe Pub's Guard has all the \nfeatures that one could ask from \na moderator bot. and many more! \n[This bot is exclusive to The Pub\n and is not found elsewhere]**_"
        file, err := os.Open("guard.webp")
        if err != nil {
            panic(err.Error())
        }
        msg := &discordgo.MessageSend{
            Embed:&discordgo.MessageEmbed{
                Title:Title,
                Color:0x00ffff,
                Description:Description,
                Image:&discordgo.MessageEmbedImage{
                    URL:"attachment://" + file.Name(),
                },
            },
            Files: []*discordgo.File{
                &discordgo.File{
                    Name:file.Name(),
                    Reader:file,
                },
            },
        }
        session.ChannelMessageSendComplex(message.ChannelID, msg)
        file.Close()
    } else if instructions[1] == "help" {
        embed := &discordgo.MessageEmbed{
            Title:"Comamnds.",
            Description:"1. `#guard mute <member> <reason>` : Mutes a member. \n 2. `#guard unmute <member>`: Unmutes a member. \n 3. `#guard clear <number>` : Deletes the specified number of messages [deletes 100 by default]. \n 4. `#guard kick <member> <reason>` : Kicks a member. \n 5. `#guard ban <member> <reason>` : Bans a member. \n 6. `#guard unban <id>` : Unbans a member.",
            Color:0x5865F2,
            Image:&discordgo.MessageEmbedImage{URL:"attachment://" + "commands.png",},
        }
        file, err := os.Open("commands.png")
        if err != nil {
            panic(err.Error())
        }
        files := []*discordgo.File{&discordgo.File{Name:file.Name(), Reader:file}}
        msg := &discordgo.MessageSend{
            Embed:embed,
            Files:files,
        }
        session.ChannelMessageSendComplex(message.ChannelID, msg)
    }
}

func Permissions(session *discordgo.Session, guild string, user string, permission int64) bool {
    member, err := session.State.Member(guild, user)
	if err != nil {
		if member, err = session.GuildMember(guild, user); err != nil {
			return false 
		}
	}
	for _, roleID := range member.Roles {
		role, err := session.State.Role(guild, roleID)
		if err != nil {
			return false 
		}
		if role.Permissions&permission != 0 || role.Permissions&discordgo.PermissionAdministrator != 0 {
			return true
		}
	}

	return false
}

func verification(session *discordgo.Session, member *discordgo.GuildMemberAdd) {
    fmt.Println(member.Mention())
    channel, err := session.UserChannelCreate(member.User.ID)
    fileName := "guard.webp"
    f, err := os.Open(fileName)
    if err != nil {
        panic(err.Error())
    }
    ms := &discordgo.MessageSend{
        Embed: &discordgo.MessageEmbed{
            Title:"The Guard",
            Description:"You have joined the pub.\n" + "**Are you a volarant or League of Legends smurf?** (yes/no)",
            Image: &discordgo.MessageEmbedImage{
                URL: "attachment://" + fileName,
            },
        },
        Files: []*discordgo.File{
            &discordgo.File{
                Name:   fileName,
                Reader: f,
            },
        },
    }
    _, err = session.ChannelMessageSendComplex(channel.ID, ms)
    if err != nil {
        panic(err.Error())
    }
    f.Close()
}
