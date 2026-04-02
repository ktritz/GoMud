
var sadSubjects = ["sad", "sadness", "point", "belong", "matter", "important", "quest", "help"];
var returnSubjects = ["learned", "recolor", "bittersweet", "tag", "station", "found", "discovered"];

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    // Check for quest 5 return conversation
    if ( user.HasQuest("5-return") && !user.HasQuest("5-end") ) {
        returnMatch = UtilFindMatchIn(eventDetails.askText, returnSubjects);
        if ( returnMatch.found ) {
            mob.Command("emote looks up slowly, eyes widening behind her glasses.");
            mob.Command("say Wait... sad memories can become... happy ones?", 2);
            mob.Command("say And happy ones can become... bittersweet?", 4);
            mob.Command("say So I'm not just making things worse?", 6);
            mob.Command("emote 's eyes fill with tears - but these are different tears.", 8);
            mob.Command("say That... actually makes me feel a lot better.", 10);
            mob.Command("say Sadness helps Riley process things. It lets other people know she needs help.", 12);
            mob.Command("say Without me... she'd just bottle everything up inside.", 14);
            mob.Command("emote gives a small, genuine smile.", 16);
            mob.Command("say Thank you. Really. You helped me understand something important.", 18);
            user.GiveQuest("5-end");
            return true;
        }

        mob.Command("say You went to the Tag Station? What did you find out?");
        mob.Command("say I've been sitting here thinking about it... probably overthinking it... that's kind of what I do.", 3);
        return true;
    }

    match = UtilFindMatchIn(eventDetails.askText, sadSubjects);
    if ( match.found ) {

        if ( user.HasQuest("5-end") ) {
            mob.Command("say I know my purpose now. Sadness is important. Thank you for helping me see that.");
            mob.Command("emote gives a small, warm smile.", 2);
            return true;
        }

        if ( user.HasQuest("5-tag") ) {
            mob.Command("say Did you visit the Tag Station yet? It's in Long Term Memory...");
            mob.Command("say I hope it's not too far. I'd go myself but... I'd probably just slow everyone down.", 3);
            return true;
        }

        if ( user.HasQuest("5-start") ) {
            mob.Command("say Did you visit the Tag Station yet? It's in Long Term Memory...");
            mob.Command("say I keep thinking about what they do there. Recoloring memories... do they make sad ones happy? Or happy ones sad?", 3);
            mob.Command("say Sorry, I'm probably bothering you with all these questions.", 6);
            return true;
        }

        if ( !user.HasQuest("5-start") ) {
            mob.Command("emote sighs deeply, staring at the floor.");
            mob.Command("say I just... I don't think I really do anything useful here.", 2);
            mob.Command("say Joy handles all the good stuff. She makes Riley happy. That's important.", 4);
            mob.Command("say I just make Riley cry.", 6);
            mob.Command("emote adjusts her glasses nervously.", 7);
            mob.Command("say Maybe... maybe you could go to the Emotional Tag Station in Long Term Memory?", 9);
            mob.Command("say I heard they recolor memories there. Change their emotional tags.", 11);
            mob.Command("say Maybe seeing how it works would help me understand... what I'm even for.", 13);
            mob.Command("say You probably don't want to do that though. It's okay. Nobody really wants to hang around sadness.", 15);
            user.GiveQuest("5-start");
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
        mob.Command("say Oh... you're giving me gold? That's... *sniff* ...that's really nice of you.");
        return true;
    }

    if ( eventDetails.item ) {
        mob.Command("say Oh... thank you. You didn't have to do that. Nobody ever gives me things.");
        mob.Command("emote holds the item close, looking like she might cry.", 2);
        mob.Command("give !" + String(eventDetails.item.ItemId) + " @" + String(eventDetails.sourceId), 4);
        return true;
    }

    return false;
}

var IDLE_LINES = [
    "emote sighs softly.",
    "say I wonder if Riley remembers the time she cried watching that movie... that was me. I did that.",
    "emote sits in the corner, hugging her knees.",
    "say Sometimes I touch a happy memory by accident and it turns blue. Joy doesn't like that.",
    "emote stares out the window at the Islands of Personality.",
    "say Do you ever feel like you don't quite fit in? ...Yeah, me neither. That would be sad. Oh wait.",
    "say I read this really sad poem once. It made me feel... understood. Is that weird?",
    "emote slowly flips through a scrapbook, sniffling quietly.",
    "say Rain is just the sky being sad. I think that's beautiful.",
];

function onIdle(mob, room) {

    round = UtilGetRoundNumber();

    if ( round % 5 != 0 ) {
        return false;
    }

    randNum = UtilDiceRoll(1, IDLE_LINES.length) - 1;
    mob.Command(IDLE_LINES[randNum]);
    return true;
}
