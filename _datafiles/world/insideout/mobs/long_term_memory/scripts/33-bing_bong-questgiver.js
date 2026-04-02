
var rocketSubjects = ["rocket", "fuel", "wagon", "song", "fly", "quest", "help", "adventure"];

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    match = UtilFindMatchIn(eventDetails.askText, rocketSubjects);
    if ( match.found ) {

        if ( user.HasQuest("2-end") ) {
            mob.Command("say We flew together! Just like me and Riley used to! That was the best day I've had in... well... a really long time.");
            mob.Command("emote quietly hums the rocket song.", 3);
            return true;
        }

        if ( user.HasQuest("2-start") ) {
            mob.Command("say Any luck finding the fuel? Try the Cookie Factory in Imagination Land!");
            mob.Command("say Or maybe near the French Fry Forest? I can't remember exactly where I left it.", 3);
            mob.Command("say My memory isn't what it used to be. *laughs* Get it? Memory? Because we're IN the memories? ...Yeah, Riley used to laugh at that one.", 6);
            return true;
        }

        if ( !user.HasQuest("2-start") ) {
            mob.Command("emote 's eyes light up and his whole cotton-candy body perks up.");
            mob.Command("say You... you want to help ME?", 2);
            mob.Command("emote wipes a candy tear from his eye.", 3);
            mob.Command("say Nobody's asked me that in a long time.", 5);
            mob.Command("say See, my rocket wagon needs fuel to fly - it runs on song!", 7);
            mob.Command("say But the special kind of fuel, I left it somewhere in Imagination Land.", 9);
            mob.Command("say Cookie Factory maybe? Or near the French Fry Forest?", 11);
            mob.Command("say If you could find it, we could fly together!", 13);
            mob.Command("emote 's voice gets quiet.", 14);
            mob.Command("say Just like me and Riley used to...", 16);
            mob.Command("say She'd climb in the wagon and we'd sing the song - 'Who's your friend who likes to play?' - and we'd go SO high!", 18);
            mob.Command("say But that was a long time ago. She doesn't come here anymore.", 20);
            user.GiveQuest("2-start");
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
        mob.Command("say Gold? I don't really use gold. My tears are made of candy though! Want one?");
        return true;
    }

    if ( eventDetails.item ) {

        // Quest 2 - Rocket Fuel (item 12)
        if ( eventDetails.item.ItemId == 12 ) {

            if ( user.HasQuest("2-start") ) {
                mob.Command("emote stares at the rocket fuel, trembling.");
                mob.Command("say You... you actually found it!", 2);
                mob.Command("emote 's candy tears stream down his face in big colorful drops.", 4);
                mob.Command("say This is the nicest thing anyone's done for me since... since Riley.", 6);
                mob.Command("say Here, take my candy cane. It's the least I can do.", 8);
                mob.Command("emote carefully places the fuel in the rocket wagon, hands shaking.", 10);
                mob.Command("say It's full. It's actually full again.", 12);
                mob.Command("emote quietly sings, voice cracking with emotion.", 14);
                mob.Command("say Who's your friend who likes to play? Bing Bong, Bing Bong...", 16);
                mob.Command("say His rocket makes you yell hooray! Bing Bong, Bing Bong...", 18);
                mob.Command("say Who's the best in every way and wants to sing this song to say...", 20);
                mob.Command("say Bing Bong! Bing Bong!", 22);
                mob.Command("emote sniffles and wipes his face, leaving candy smears everywhere.", 24);
                mob.Command("say Thank you. Really. You made an old imaginary friend very happy today.", 26);
                user.GiveQuest("2-end");
                return true;
            }

            mob.Command("say Is that... rocket fuel? I don't think I asked for any, but it sure brings back memories!");
            mob.Command("give !" + String(eventDetails.item.ItemId) + " @" + String(eventDetails.sourceId));
            return true;
        }

        mob.Command("say For me? Aw, you're sweet. But I'm really just looking for my rocket fuel.");
        mob.Command("give !" + String(eventDetails.item.ItemId) + " @" + String(eventDetails.sourceId), 2);
        return true;
    }

    return false;
}

var IDLE_LINES = [
    "say I wonder what Riley's doing right now...",
    "emote quietly sings the rocket song to himself.",
    "say She used to imagine me every single day, you know.",
    "emote pats the side of his rocket wagon gently.",
    "say I'm part cat, part elephant, part dolphin. Riley made me up when she was three!",
    "say My tears are candy! Watch! *sniff* ...See? Delicious AND sad!",
    "emote traces a heart shape in the dust and writes 'BingBong + Riley' inside it.",
    "say This place used to be so busy. Memories everywhere, all nice and organized. Now they just keep fading...",
    "say If you hum a really happy song, the memories glow brighter. Try it sometime!",
    "emote sits in his rocket wagon and pretends to fly, making quiet whooshing sounds.",
    "say Take her to the moon for me. Okay?",
];

function onIdle(mob, room) {

    round = UtilGetRoundNumber();

    if ( round % 4 != 0 ) {
        return false;
    }

    missingQuestUsers = room.MissingQuest("2-start");
    if ( missingQuestUsers.length > 0 ) {
        mob.Command("say Hey there, friend! You wouldn't happen to know where I could find some rocket fuel, would you? My wagon hasn't flown in ages...");
        return true;
    }

    randNum = UtilDiceRoll(1, IDLE_LINES.length) - 1;
    mob.Command(IDLE_LINES[randNum]);
    return true;
}
