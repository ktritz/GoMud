
var fearSubjects = ["fear", "subconscious", "scary", "noise", "disturbance", "quest", "help", "jangles"];
var returnSubjects = ["safe", "defeated", "jangles", "done", "beaten", "cleared", "back", "report"];

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    // Check for quest 6 return conversation
    if ( user.HasQuest("6-defeat") && !user.HasQuest("6-end") ) {
        returnMatch = UtilFindMatchIn(eventDetails.askText, returnSubjects);
        if ( returnMatch.found ) {
            mob.Command("emote 's eyes go wide and his jaw drops.");
            mob.Command("say The Subconscious is safe again?", 2);
            mob.Command("say I can sleep tonight?", 3);
            mob.Command("say Well, I don't actually sleep, but the CONCEPT of sleeping?", 5);
            mob.Command("emote hyperventilates into a paper bag for a moment.", 7);
            mob.Command("say Thank you. You're much braver than I'll ever be.", 9);
            mob.Command("say And that's saying something because I'm terrified of literally everything.", 11);
            mob.Command("say Stairs, the dark, broccoli, that weird noise the fridge makes, dogs that are too friendly...", 13);
            mob.Command("say The point is, you did what I couldn't. And Riley is safer because of it.", 15);
            user.GiveQuest("6-end");
            return true;
        }

        mob.Command("say You BEAT him?!");
        mob.Command("emote hyperventilates briefly.", 2);
        mob.Command("say Oh thank goodness. I was already writing your eulogy just in case. It was very touching, by the way.", 4);
        mob.Command("say So the Subconscious is safe now? Tell me it's safe!", 6);
        return true;
    }

    match = UtilFindMatchIn(eventDetails.askText, fearSubjects);
    if ( match.found ) {

        if ( user.HasQuest("6-end") ) {
            mob.Command("say The Subconscious is secure, thanks to you. I've only had three panic attacks today instead of the usual twelve!");
            return true;
        }

        if ( user.HasQuest("6-investigate") ) {
            mob.Command("say Be careful down there! Jangles is NOT a friendly clown.");
            mob.Command("say I have a list of 47 ways he could hurt you and that's just off the top of my head.", 2);
            mob.Command("say I've also prepared a risk assessment matrix, a danger probability chart, and a last will and testament template. Want any of those?", 5);
            return true;
        }

        if ( user.HasQuest("6-start") ) {
            mob.Command("say You're heading to the Subconscious? Oh boy. Oh no. Oh boy oh no.");
            mob.Command("say Look for Jangles in the containment area. And whatever you do, don't let him honk at you!", 3);
            return true;
        }

        if ( !user.HasQuest("6-start") ) {
            mob.Command("emote jumps at the sound of his own voice.");
            mob.Command("say AAGH! Oh, it's you. Sorry. Sorry, I'm a little on edge.", 2);
            mob.Command("say Something is VERY wrong in the Subconscious!", 4);
            mob.Command("say I keep hearing noises - BAD noises. Like... honking? And cackling?", 6);
            mob.Command("emote shivers uncontrollably.", 7);
            mob.Command("say I think Jangles has gotten out of his cell!", 9);
            mob.Command("say Jangles the Clown! Riley's childhood fear! Big shoes, creepy smile, TERRIFYING honking!", 11);
            mob.Command("say Could you PLEASE go check? I would go myself but... you know... survival instinct.", 13);
            mob.Command("say Also I'm scared. Really, really scared. Did I mention I'm scared?", 15);
            user.GiveQuest("6-start");
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
        mob.Command("say Gold?! Is it cursed? It's probably cursed. ...Thank you though.");
        return true;
    }

    if ( eventDetails.item ) {
        mob.Command("say What is this?! Is it dangerous? It looks dangerous. Everything looks dangerous to me.");
        mob.Command("give !" + String(eventDetails.item.ItemId) + " @" + String(eventDetails.sourceId), 2);
        return true;
    }

    return false;
}

var IDLE_LINES = [
    "emote nervously checks behind the console for the third time.",
    "say Did you hear that? ...No? Just me? Okay. That's somehow worse.",
    "emote flinches at nothing in particular.",
    "say I've cataloged 2,347 things that could go wrong today. And that's just before lunch.",
    "say You know what's scary? EVERYTHING. That's what.",
    "emote clutches his clipboard like a security blanket.",
    "say I made Riley avoid that stray dog yesterday. You're WELCOME.",
    "say The probability of something terrible happening is never zero. Think about THAT.",
    "emote peeks out the window, then immediately ducks back down.",
    "say Did anyone else feel that earthquake? No? Micro-tremor? Pre-earthquake? ...Just gas?",
];

function onIdle(mob, room) {

    round = UtilGetRoundNumber();

    if ( round % 4 != 0 ) {
        return false;
    }

    missingQuestUsers = room.MissingQuest("6-start");
    if ( missingQuestUsers.length > 0 ) {
        mob.Command("say I keep hearing weird noises from the Subconscious... like honking... and maniacal laughter. This is NOT good.");
        return true;
    }

    randNum = UtilDiceRoll(1, IDLE_LINES.length) - 1;
    mob.Command(IDLE_LINES[randNum]);
    return true;
}
