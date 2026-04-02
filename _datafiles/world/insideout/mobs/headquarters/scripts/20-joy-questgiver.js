
var memorySubjects = ["memory", "core", "missing", "vault", "quest", "help"];
var anxietySubjects = ["anxiety", "overthrow", "vault", "locked", "emotions", "back of mind"];
var voidSubjects = ["void", "corruption", "dump", "memories", "dark"];

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    // Quest 10 - Into the Void
    voidMatch = UtilFindMatchIn(eventDetails.askText, voidSubjects);
    if ( voidMatch.found ) {

        if ( user.HasQuest("10-end") ) {
            mob.Command("say We saved the memories! Everything is bright and beautiful again!");
            return true;
        }

        if ( user.HasQuest("10-return") ) {
            mob.Command("say You made it back! Tell me everything - did you find what was causing the corruption?");
            return true;
        }

        if ( user.HasQuest("10-start") ) {
            mob.Command("say Please be careful down in the Memory Dump. Whatever is corrupting those memories... it feels really, really wrong.");
            return true;
        }

        if ( user.HasQuest("8-end") && !user.HasQuest("10-start") ) {
            mob.Command("emote 's smile fades, and for a moment she looks genuinely serious.");
            mob.Command("say Something is wrong in the Memory Dump.", 2);
            mob.Command("say Memories are being corrupted faster than ever. The workers down there can barely keep up.", 4);
            mob.Command("say I think... I think something down there is pulling everything in. Like a dark gravity.", 6);
            mob.Command("say I know I'm usually all about looking on the bright side, but this scares me.", 8);
            mob.Command("say Can you go find out what it is? For Riley?", 10);
            user.GiveQuest("10-start");
            return true;
        }

        return false;
    }

    // Quest 8 - Anxiety's Overthrow
    anxietyMatch = UtilFindMatchIn(eventDetails.askText, anxietySubjects);
    if ( anxietyMatch.found ) {

        if ( user.HasQuest("8-end") ) {
            mob.Command("say You saved us all! Headquarters is back to normal, and Anxiety is... well, she's still anxious, but at least she's not running the show anymore.");
            return true;
        }

        if ( user.HasQuest("8-start") ) {
            mob.Command("say Please hurry! Find the suppression key in the Belief System.");
            mob.Command("say Some of us are trapped in the Back of the Mind and it's SO dark in there!", 2);
            return true;
        }

        if ( user.HasQuest("3-end") && !user.HasQuest("8-start") ) {
            mob.Command("emote 's glow flickers, and her eyes well up with tears.");
            mob.Command("say Something terrible has happened.", 2);
            mob.Command("say Anxiety has locked us out! She took over the console and...", 4);
            mob.Command("emote 's voice breaks.", 5);
            mob.Command("say ...she put some of us in the Back of the Mind. Disgust, Fear... they're trapped!", 7);
            mob.Command("say We need someone brave to get the suppression key from the Belief System and free everyone.", 9);
            mob.Command("say The Belief System is deep in Long Term Memory. The key should be somewhere inside.", 11);
            mob.Command("say Please... we can't let Anxiety run things alone. Riley needs ALL of us.", 13);
            user.GiveQuest("8-start");
            return true;
        }

        return false;
    }

    // Quest 3 - Joy's Missing Core Memory
    memoryMatch = UtilFindMatchIn(eventDetails.askText, memorySubjects);
    if ( memoryMatch.found ) {

        if ( user.HasQuest("3-end") ) {
            mob.Command("say The core memory is safe and sound back in the vault! You're the best!");
            return true;
        }

        if ( user.HasQuest("3-search") ) {
            mob.Command("say Oh, you found something? Show me! Or better yet, give it to me so I can check if it's the right one!");
            return true;
        }

        if ( user.HasQuest("3-start") ) {
            mob.Command("say Any luck finding it? Try the fading memory aisles - those Memory Leeches are always causing trouble!");
            mob.Command("say The core memory is golden and glowy - you can't miss it. Well, I mean, someone DID miss it, because it's gone, but you know what I mean!", 3);
            return true;
        }

        if ( !user.HasQuest("3-start") ) {
            mob.Command("emote gasps and rushes over.");
            mob.Command("say Oh, thank goodness someone's here!", 2);
            mob.Command("say A golden core memory has gone missing from the vault! One of Riley's most important ones!", 4);
            mob.Command("say I just don't understand how it could have disappeared! It was RIGHT there on the shelf!", 6);
            mob.Command("say Can you check Long Term Memory for me? Please?", 8);
            mob.Command("say The Memory Leeches down there are always nibbling at things they shouldn't. Maybe one of them snagged it!", 10);
            user.GiveQuest("3-start");
            return true;
        }
    }

    return false;
}

function onGive(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    if ( eventDetails.gold > 0 ) {
        mob.Command("say Aw, that's so generous! But I don't really need gold - I run on happiness!");
        return true;
    }

    if ( eventDetails.item ) {

        // Quest 3 - Core Memory Gold (item 8)
        if ( eventDetails.item.ItemId == 8 ) {

            if ( user.HasQuest("3-search") ) {
                mob.Command("emote 's entire body lights up like a supernova.");
                mob.Command("say OH MY GOSH!", 2);
                mob.Command("say You found it! You actually found it!", 3);
                mob.Command("emote bounces up and down, leaving little trails of golden light.", 4);
                mob.Command("say Thank you thank you THANK YOU!", 6);
                mob.Command("say Riley's memory is safe! This is the BEST day!", 8);
                mob.Command("say Here, take this - you deserve something special for being so amazing!", 10);
                user.GiveQuest("3-end");
                return true;
            }

            mob.Command("say Ooh, a golden memory! But... hmm, I don't think I asked for this one. Hold onto it though, it's pretty!");
            mob.Command("give !" + String(eventDetails.item.ItemId) + " @" + String(eventDetails.sourceId));
            return true;
        }

        // Quest 10 - Void completion token
        if ( user.HasQuest("10-return") ) {
            mob.Command("emote clasps her hands together, tears of joy streaming down her face.");
            mob.Command("say You did it! You actually went into the void and came back!", 2);
            mob.Command("say The Memory Dump is safe again! Riley's memories are SAFE!", 4);
            mob.Command("emote does a little spinning dance, showering golden sparkles everywhere.", 6);
            mob.Command("say I knew you could do it! I believed in you the WHOLE time!", 8);
            mob.Command("say This calls for a CELEBRATION! Group hug, everyone! GROUP HUG!", 10);
            user.GiveQuest("10-end");
            return true;
        }

        // Wrong item
        mob.Command("say That's sweet, but it's not the core memory I'm looking for.");
        mob.Command("give !" + String(eventDetails.item.ItemId) + " @" + String(eventDetails.sourceId));
        return true;
    }

    return false;
}

var IDLE_LINES = [
    "emote hums a cheerful little tune, glowing softly.",
    "say Every day is a great day when you choose to see the bright side!",
    "emote twirls around, leaving a trail of golden sparkles.",
    "say I love being in charge of Riley's happiness. It's literally the best job ever!",
    "emote adjusts the core memories in the vault, making sure they're all perfectly aligned.",
    "say Did you know that Riley smiled 47 times yesterday? I counted!",
    "emote bounces on her toes, radiating warmth.",
];

function onIdle(mob, room) {

    round = UtilGetRoundNumber();

    if ( round % 4 != 0 ) {
        return false;
    }

    missingQuestUsers = room.MissingQuest("3-start");
    if ( missingQuestUsers.length > 0 ) {
        mob.Command("say Has anyone seen a golden core memory? I'm really worried - one of Riley's most important memories is missing!");
        return true;
    }

    randNum = UtilDiceRoll(1, IDLE_LINES.length) - 1;
    mob.Command(IDLE_LINES[randNum]);
    return true;
}
