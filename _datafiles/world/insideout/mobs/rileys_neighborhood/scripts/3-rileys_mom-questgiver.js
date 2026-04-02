const ratNouns = ["rat", "rats", "basement", "scratching", "noise", "problem"];

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    // Quest not started yet
    if ( !user.HasQuest("11-start") ) {
        startMatch = UtilFindMatchIn(eventDetails.askText, ratNouns);
        if ( startMatch.found ) {
            mob.Command("say Oh honey, I keep hearing scratching down in the basement. I think we have rats!");
            mob.Command("say I'd go down there myself but... you know. Could you take care of it?");
            mob.Command("say If you can get rid of 5 of them, I'll feel a lot better.");
            mob.Command("emote shudders at the thought of rats.");

            user.GiveQuest("11-start");
            return true;
        }
        return false;
    }

    // Quest in progress — check rat kills
    if ( user.HasQuest("11-start") && !user.HasQuest("11-end") ) {
        ratkillCt = user.GetRaceKills("vermin");

        if ( ratkillCt >= 5 ) {
            mob.Command("say Oh thank goodness! No more scratching?");
            mob.Command("say You're the best. Here, take this. You earned it.");
            mob.Command("emote gives you a big relieved hug.");

            user.GiveQuest("11-end");
            return true;
        }

        mob.Command("say How's it going down there? You've gotten " + String(ratkillCt) + " so far. There might be more hiding behind the boxes.");
        return true;
    }

    // Quest complete
    if ( user.HasQuest("11-end") ) {
        ratMatch = UtilFindMatchIn(eventDetails.askText, ratNouns);
        if ( ratMatch.found ) {
            mob.Command("say I haven't heard any scratching lately, thank goodness. You really took care of it!");
            return true;
        }
    }

    return false;
}
