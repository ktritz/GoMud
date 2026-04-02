var homeworkSubjects = ["homework", "assignment", "class", "english", "quest", "help", "paper"];
var quizSubjects = ["quiz", "test", "pop quiz", "exam"];

// English Quiz - Multiple choice, player says A/B/C or the answer word
var quizQuestions = [
    {
        question: "What word is a synonym for 'happy'? <ansi fg=\"yellow\">A)</ansi> elated, <ansi fg=\"yellow\">B)</ansi> hungry, <ansi fg=\"yellow\">C)</ansi> heavy",
        answers: ["a", "elated"],
        step: "q1",
        nextStep: "12-q2"
    },
    {
        question: "What does the word 'reluctant' mean? <ansi fg=\"yellow\">A)</ansi> eager, <ansi fg=\"yellow\">B)</ansi> unwilling, <ansi fg=\"yellow\">C)</ansi> tired",
        answers: ["b", "unwilling"],
        step: "q2",
        nextStep: "12-q3"
    },
    {
        question: "In a story, the 'climax' is: <ansi fg=\"yellow\">A)</ansi> the beginning, <ansi fg=\"yellow\">B)</ansi> the setting, <ansi fg=\"yellow\">C)</ansi> the turning point",
        answers: ["c", "turning point"],
        step: "q3",
        nextStep: "12-end"
    }
];

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    // Handle homework quest (quest 1)
    match = UtilFindMatchIn(eventDetails.askText, homeworkSubjects);
    if ( match.found ) {

        if ( user.HasQuest("1-end") ) {
            mob.Command("say Glad you got that sorted out! Remember, next assignment is due Friday.");
            return true;
        }

        if ( user.HasQuest("1-start") ) {
            mob.Command("say Still looking for that homework? Check around the park and maybe the neighbor's yards.");
            return true;
        }

        if ( !user.HasQuest("1-start") ) {
            mob.Command("emote looks up from her desk and adjusts her reading glasses.");
            mob.Command("say Oh, you're looking for your homework?", 2);
            mob.Command("say I noticed you didn't turn it in this morning. That's not like you.", 4);
            mob.Command("say The wind was really something today - I saw papers blowing everywhere near the park.", 6);
            mob.Command("say If you can find it and bring it back, I'll still give you full credit.", 8);
            mob.Command("say Try retracing your steps from the park.", 10);
            user.GiveQuest("1-start");
            return true;
        }
    }

    // Handle English quiz (quest 12) - gated behind homework quest
    quizMatch = UtilFindMatchIn(eventDetails.askText, quizSubjects);
    if ( quizMatch.found ) {

        if ( !user.HasQuest("1-end") ) {
            mob.Command("say Let's get your homework sorted out first before we talk about quizzes, okay?");
            return true;
        }

        if ( user.HasQuest("12-end") ) {
            mob.Command("say You already aced my quiz! Gold star for you. Maybe try Mr. Chen's math quiz next.");
            return true;
        }

        if ( !user.HasQuest("12-start") ) {
            mob.Command("emote puts down her red pen and looks at you over her glasses.");
            mob.Command("say Oh, you want to try the pop quiz? Three questions. Multiple choice.", 2);
            mob.Command("say Just <ansi fg=\"command\">say</ansi> the letter or the answer. Ready?", 4);
            mob.Command("say Question 1: " + quizQuestions[0].question, 6);
            user.GiveQuest("12-start");
            user.GiveQuest("12-q1");
            return true;
        }

        // Repeat current question
        for ( i = 0; i < quizQuestions.length; i++ ) {
            if ( user.HasQuest("12-" + quizQuestions[i].step) && !user.HasQuest(quizQuestions[i].nextStep) ) {
                mob.Command("say " + quizQuestions[i].question);
                return true;
            }
        }
    }

    return false;
}

function onSay(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    if ( !user.HasQuest("12-start") || user.HasQuest("12-end") ) {
        return false;
    }

    answer = eventDetails.msg.toLowerCase().trim();

    for ( i = 0; i < quizQuestions.length; i++ ) {
        q = quizQuestions[i];

        if ( !user.HasQuest("12-" + q.step) || user.HasQuest(q.nextStep) ) {
            continue;
        }

        correct = false;
        for ( a = 0; a < q.answers.length; a++ ) {
            if ( answer == q.answers[a] || answer.indexOf(q.answers[a]) >= 0 ) {
                correct = true;
                break;
            }
        }

        if ( correct ) {
            user.GiveQuest(q.nextStep);

            if ( q.nextStep == "12-end" ) {
                mob.Command("say Correct! Perfect score! You really know your English.");
                mob.Command("emote marks an A+ on her clipboard.", 2);
                mob.Command("say If you're feeling confident, Mr. Chen in Classroom 202 has a math quiz.", 4);
            } else {
                nextIdx = i + 1;
                mob.Command("say Correct!");
                mob.Command("say Question " + String(nextIdx + 1) + ": " + quizQuestions[nextIdx].question, 2);
            }
            return true;
        } else {
            mob.Command("say Not quite. Try again! Remember, just say the letter or the answer.");
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
        mob.Command("say Teachers can't accept gifts from students. School policy!");
        mob.Command("give " + String(eventDetails.gold) + " gold @" + String(eventDetails.sourceId), 2);
        return true;
    }

    if ( eventDetails.item ) {
        if ( eventDetails.item.ItemId == 2 && user.HasQuest("1-start") ) {
            mob.Command("emote takes the crumpled paper and smooths it out.");
            mob.Command("say You found it! Full marks for persistence.", 2);
            user.GiveQuest("1-end");
            return true;
        }

        mob.Command("say Not sure what I'd do with that. Hold onto it.");
        mob.Command("give !" + String(eventDetails.item.ItemId) + " @" + String(eventDetails.sourceId), 2);
        return true;
    }

    return false;
}

function onIdle(mob, room) {

    if ( UtilGetRoundNumber() % 5 != 0 ) {
        return false;
    }

    missingQuestUsers = room.MissingQuest("12-start");
    if ( missingQuestUsers.length > 0 ) {
        mob.Command("say Anyone want to try today's pop quiz? Just <ansi fg=\"command\">ask</ansi> me about the <ansi fg=\"command\">quiz</ansi>!");
        return true;
    }

    lines = [
        "emote shuffles papers on her desk.",
        "say Reading is fundamental!",
        "emote writes on the whiteboard.",
        "emote sips from a mug that says 'World's Okayest Teacher'.",
    ];

    mob.Command(lines[UtilDiceRoll(1, lines.length) - 1]);
    return true;
}
