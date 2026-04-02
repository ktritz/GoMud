var quizSubjects = ["quiz", "test", "pop quiz", "exam", "pe", "health", "fitness", "challenge"];
var PREV_QUEST = "18-end";

var tiers = [
    {
        questId: "19",
        name: "Beginner",
        prevQuest: PREV_QUEST,
        nextTierMsg: "Not bad for a warmup! Ask me about the challenge again for the INTERMEDIATE tier!",
        challenges: [
            { prompt: "Basic pushups! Give me ten!", stat: "strength", threshold: 5,
              successMsg: "says, \"Good form! You got through it!\"",
              failMsg: "says, \"Come on, you can do better! Train that strength!\"" },
            { prompt: "Jog a lap around the gym!", stat: "speed", threshold: 5,
              successMsg: "says, \"Nice pace! You finished strong!\"",
              failMsg: "says, \"You're dragging! Work on that speed!\"" },
            { prompt: "Hold a plank for 30 seconds!", stat: "vitality", threshold: 5,
              successMsg: "says, \"Solid! You didn't even shake!\"",
              failMsg: "says, \"You folded like a lawn chair! Build that vitality!\"" }
        ]
    },
    {
        questId: "21",
        name: "Intermediate",
        prevQuest: "19-end",
        nextTierMsg: "Impressive! Ask me again for the ADVANCED tier... if you dare!",
        challenges: [
            { prompt: "One-arm pushups! Show me what you've got!", stat: "strength", threshold: 12,
              successMsg: "says, \"Now THAT takes real strength!\"",
              failMsg: "says, \"Your arm buckled! You need more strength training!\"" },
            { prompt: "Sprint drill! Wall to wall, five times, GO!", stat: "speed", threshold: 12,
              successMsg: "says, \"You're a blur! Outstanding speed!\"",
              failMsg: "says, \"Too slow by half! Speed it up!\"" },
            { prompt: "Wall sit until I say stop!", stat: "vitality", threshold: 12,
              successMsg: "says, \"Your legs are made of iron!\"",
              failMsg: "says, \"Your legs are shaking like jelly! More endurance training!\"" }
        ]
    },
    {
        questId: "22",
        name: "Advanced",
        prevQuest: "21-end",
        nextTierMsg: "You've completed ALL my challenges! You're a true athlete!",
        challenges: [
            { prompt: "Handstand pushups against the wall! This is the real deal!", stat: "strength", threshold: 20,
              successMsg: "says, \"INCREDIBLE! That takes serious power!\"",
              failMsg: "says, \"You crashed down! You need elite-level strength for this one!\"" },
            { prompt: "Agility course! Through the cones, over the hurdles, under the bars!", stat: "speed", threshold: 20,
              successMsg: "says, \"You moved like lightning through that course!\"",
              failMsg: "says, \"You clipped every hurdle! Get faster and try again!\"" },
            { prompt: "Full gym circuit! Every station, no breaks!", stat: "vitality", threshold: 20,
              successMsg: "says, \"You finished the whole circuit without stopping! BEAST MODE!\"",
              failMsg: "says, \"You gassed out halfway! Build that endurance!\"" }
        ]
    }
];

function findActiveTier(user) {
    // Find which tier the user should be working on
    for ( t = tiers.length - 1; t >= 0; t-- ) {
        if ( user.HasQuest(tiers[t].questId + "-start") ) return t;
    }
    // Find the next tier to offer
    for ( t = 0; t < tiers.length; t++ ) {
        if ( !user.HasQuest(tiers[t].questId + "-end") ) return t;
    }
    return -1; // all done
}

function onAsk(mob, room, eventDetails) {
    if ( (user = GetUser(eventDetails.sourceId)) == null ) return false;

    quizMatch = UtilFindMatchIn(eventDetails.askText, quizSubjects);
    if ( !quizMatch.found ) return false;

    if ( !user.HasQuest(PREV_QUEST) ) {
        mob.Command("say Finish Mr. Grant's history quiz first! Knowledge is power!");
        return true;
    }

    tierIdx = findActiveTier(user);
    if ( tierIdx < 0 ) {
        mob.Command("say You've beaten ALL my challenges! You're the toughest kid in this school! Try Ms. Tanaka's programming quiz in the Computer Lab!");
        return true;
    }

    tier = tiers[tierIdx];

    if ( user.HasQuest(tier.questId + "-end") ) {
        // This tier is done, offer next
        tierIdx++;
        if ( tierIdx >= tiers.length ) {
            mob.Command("say You've beaten ALL my challenges! Try Ms. Tanaka's programming quiz!");
            return true;
        }
        tier = tiers[tierIdx];
    }

    if ( !user.HasQuest(tier.questId + "-start") ) {
        mob.Command("emote blows his whistle sharply.");
        mob.Command("say " + tier.name + " PE challenges! Three physical tests!", 2);
        mob.Command("say Just <ansi fg=\"command\">say ready</ansi> to attempt each one.", 4);
        mob.Command("say Challenge 1: " + tier.challenges[0].prompt, 6);
        user.GiveQuest(tier.questId + "-start");
        user.GiveQuest(tier.questId + "-q1");
        return true;
    }

    // Show current challenge
    for ( i = 0; i < tier.challenges.length; i++ ) {
        step = "q" + String(i + 1);
        nextStep = (i < tier.challenges.length - 1) ? tier.questId + "-q" + String(i + 2) : tier.questId + "-end";
        if ( user.HasQuest(tier.questId + "-" + step) && !user.HasQuest(nextStep) ) {
            statVal = user.GetStat(tier.challenges[i].stat);
            mob.Command("say " + tier.challenges[i].prompt);
            mob.Command("say Your " + tier.challenges[i].stat + ": <ansi fg=\"yellow\">" + String(statVal) + "</ansi>. Target: <ansi fg=\"red\">" + String(tier.challenges[i].threshold) + "</ansi> (stat + d20 roll).", 2);
            mob.Command("say <ansi fg=\"command\">Say ready</ansi> to try!", 4);
            return true;
        }
    }

    return false;
}

function onSay(mob, room, eventDetails) {
    if ( (user = GetUser(eventDetails.sourceId)) == null ) return false;

    msg = eventDetails.msg.toLowerCase().trim();
    if ( msg != "ready" && msg != "go" && msg != "yes" && msg != "try" ) return false;

    tierIdx = findActiveTier(user);
    if ( tierIdx < 0 ) return false;

    tier = tiers[tierIdx];
    if ( !user.HasQuest(tier.questId + "-start") || user.HasQuest(tier.questId + "-end") ) return false;

    for ( i = 0; i < tier.challenges.length; i++ ) {
        step = "q" + String(i + 1);
        nextStep = (i < tier.challenges.length - 1) ? tier.questId + "-q" + String(i + 2) : tier.questId + "-end";
        if ( !user.HasQuest(tier.questId + "-" + step) || user.HasQuest(nextStep) ) continue;

        c = tier.challenges[i];
        statVal = user.GetStat(c.stat);
        roll = UtilDiceRoll(1, 20);
        total = statVal + roll;

        mob.Command("say Let's see what you've got!");

        if ( total >= c.threshold ) {
            SendUserMessage(user.UserId(), "Rolled <ansi fg=\"yellow\">" + String(roll) + "</ansi> + " + c.stat + " <ansi fg=\"yellow\">" + String(statVal) + "</ansi> = <ansi fg=\"green\">" + String(total) + "</ansi> vs <ansi fg=\"red\">" + String(c.threshold) + "</ansi> -- <ansi fg=\"green\">SUCCESS!</ansi>");
            mob.Command(c.successMsg, 2);
            user.GiveQuest(nextStep);

            if ( nextStep == tier.questId + "-end" ) {
                mob.Command("emote gives you a firm handshake.", 4);
                mob.Command("say " + tier.nextTierMsg, 6);
            } else {
                nextIdx = i + 1;
                mob.Command("say Challenge " + String(nextIdx + 1) + ": " + tier.challenges[nextIdx].prompt, 4);
                mob.Command("say <ansi fg=\"command\">Say ready</ansi> to try!", 6);
            }
        } else {
            SendUserMessage(user.UserId(), "Rolled <ansi fg=\"yellow\">" + String(roll) + "</ansi> + " + c.stat + " <ansi fg=\"yellow\">" + String(statVal) + "</ansi> = <ansi fg=\"red\">" + String(total) + "</ansi> vs <ansi fg=\"red\">" + String(c.threshold) + "</ansi> -- <ansi fg=\"red\">FAILED!</ansi>");
            mob.Command(c.failMsg, 2);
            mob.Command("say Train up and <ansi fg=\"command\">say ready</ansi> to try again!", 4);
        }
        return true;
    }
    return false;
}

function onIdle(mob, room) {
    if ( UtilGetRoundNumber() % 5 != 0 ) return false;
    // Check for any tier they haven't started
    for ( t = 0; t < tiers.length; t++ ) {
        if ( room.MissingQuest(tiers[t].questId + "-start").length > 0 ) {
            mob.Command("say PE challenge! <ansi fg=\"command\">Ask</ansi> me about the <ansi fg=\"command\">challenge</ansi>!");
            return true;
        }
    }
    lines = [
        "emote does pushups to prove a point nobody asked about.",
        "say HYDRATE! I cannot stress this enough!",
        "emote checks his stopwatch and shakes his head.",
    ];
    mob.Command(lines[UtilDiceRoll(1, lines.length) - 1]);
    return true;
}
