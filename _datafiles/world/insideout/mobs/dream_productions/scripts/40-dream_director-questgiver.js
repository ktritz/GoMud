
var dreamSubjects = ["dream", "nightmare", "script", "wrong", "corrupted", "quest", "help", "horror"];

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    match = UtilFindMatchIn(eventDetails.askText, dreamSubjects);
    if ( match.found ) {

        if ( user.HasQuest("4-end") ) {
            mob.Command("say We're back on schedule thanks to you, darling! Tonight's dream is a musical about flying pizza. It's going to be MAGNIFICENT.");
            return true;
        }

        if ( user.HasQuest("4-script") ) {
            mob.Command("say The corrupted script has got to be in the Script Archive somewhere! Find it and bring it to me.");
            mob.Command("say Every minute it's out there, more nightmares are spawning from the terrible rewrites!", 3);
            return true;
        }

        if ( user.HasQuest("4-clear") ) {
            mob.Command("say Good, you're thinning out the nightmares! Now find that corrupted script - try the Script Archive!");
            mob.Command("say It's the root of all this chaos. Without it, we can get production back on track!", 3);
            return true;
        }

        if ( user.HasQuest("4-start") ) {
            mob.Command("say Still working on clearing out Soundstage B? Those nightmare creatures won't clean themselves up!");
            mob.Command("say And we have a DEADLINE, people! Riley needs to dream about showing up to school in her underwear by Thursday!", 3);
            return true;
        }

        if ( !user.HasQuest("4-start") ) {
            mob.Command("emote throws his megaphone on the ground in frustration.");
            mob.Command("say CUT! CUT CUT CUT!", 2);
            mob.Command("emote picks the megaphone back up, then throws it down again.", 3);
            mob.Command("say Everything's gone WRONG!", 5);
            mob.Command("say Soundstage B was supposed to be a romantic comedy - boy meets girl, they share a sandwich, very heartwarming stuff.", 7);
            mob.Command("say But some MANIAC rewrote the script and now it's full of nightmares!", 9);
            mob.Command("say The creatures have spread into the Nightmare Wing! My crew is terrified! Carlos in props hasn't stopped screaming since Tuesday!", 11);
            mob.Command("say I need someone to clear them out AND find that corrupted script!", 13);
            mob.Command("say This is a PRODUCTION, people! We are on a SCHEDULE!", 15);
            mob.Command("say Riley can't have nightmares every single night - her parents will start asking questions!", 17);
            user.GiveQuest("4-start");
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
        mob.Command("say We don't take bribes in Dream Productions! ...Well, not SMALL bribes anyway.");
        return true;
    }

    if ( eventDetails.item ) {

        // Quest 4 - Dream Script (item 6)
        if ( eventDetails.item.ItemId == 6 ) {

            if ( user.HasQuest("4-script") || user.HasQuest("4-clear") ) {
                mob.Command("emote snatches the script and holds it up to the light.");
                mob.Command("say The script! Let me see...", 2);
                mob.Command("emote flips through the pages rapidly.", 3);
                mob.Command("say 'AND THEN EVERYTHING IS SCARY FOREVER.' That's on page one.", 5);
                mob.Command("say Page two: 'MORE SCARY. TEETH EVERYWHERE.'", 7);
                mob.Command("say Page three: 'THE TEETH HAVE TEETH.'", 9);
                mob.Command("say Oh no wonder - someone wrote 'AND THEN EVERYTHING IS SCARY FOREVER' on every single page. Amateur hour.", 11);
                mob.Command("say No character development, no narrative arc, just TEETH. This is HACK writing!", 13);
                mob.Command("say Well, we can fix this. I'll have the writers do a full rewrite by morning.", 15);
                mob.Command("say Thank you, darling, you're a STAR.", 17);
                mob.Command("say Take this beret - you're officially part of the crew. Welcome to Dream Productions!", 19);
                mob.Command("emote pins a tiny 'CREW' badge on you.", 21);
                user.GiveQuest("4-end");
                return true;
            }

            mob.Command("say A dream script? I have plenty of those, darling. I need the CORRUPTED one from Soundstage B.");
            mob.Command("give !" + String(eventDetails.item.ItemId) + " @" + String(eventDetails.sourceId), 2);
            return true;
        }

        mob.Command("say What is this? This isn't in the script! NOTHING happens that isn't in the script!");
        mob.Command("give !" + String(eventDetails.item.ItemId) + " @" + String(eventDetails.sourceId), 2);
        return true;
    }

    return false;
}

var IDLE_LINES = [
    "emote yells into a megaphone at nobody in particular.",
    "say QUIET ON THE SET! ...We're not rolling? Fine. LOUD ON THE SET!",
    "say Tonight's dream: Riley shows up to school and all her teeth fall out. Classic!",
    "emote reviews storyboards, crossing things out aggressively with a red pen.",
    "say Where is my COFFEE? A director cannot direct without COFFEE!",
    "say Last night's dream about the talking pizza got a 94% approval rating from the subconscious. I'm basically a genius.",
    "emote adjusts his beret and scarf for the fourteenth time.",
    "say DO NOT touch the nightmare props! Last intern who did that didn't sleep for a week. Ironic, really.",
    "say Art is SUFFERING, people! And by people I mean mostly me!",
    "emote dramatically reads from a script, acting out all the parts himself.",
];

function onIdle(mob, room) {

    round = UtilGetRoundNumber();

    if ( round % 4 != 0 ) {
        return false;
    }

    missingQuestUsers = room.MissingQuest("4-start");
    if ( missingQuestUsers.length > 0 ) {
        mob.Command("say Someone PLEASE help me with this nightmare situation on Soundstage B! This is a CRISIS!");
        return true;
    }

    randNum = UtilDiceRoll(1, IDLE_LINES.length) - 1;
    mob.Command(IDLE_LINES[randNum]);
    return true;
}
