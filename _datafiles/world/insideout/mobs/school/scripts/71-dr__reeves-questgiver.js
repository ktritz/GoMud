var quizSubjects = ["quiz", "test", "pop quiz", "exam", "science"];
var QUEST_ID = "14";
var PREV_QUEST = "13-end"; // Must complete math first

var quizQuestions = [
    {
        question: "What gas do plants absorb from the air? <ansi fg=\"yellow\">A)</ansi> oxygen, <ansi fg=\"yellow\">B)</ansi> carbon dioxide, <ansi fg=\"yellow\">C)</ansi> nitrogen",
        answers: ["b", "carbon dioxide"],
        step: "q1",
        nextStep: QUEST_ID + "-q2"
    },
    {
        question: "What force keeps us on the ground? <ansi fg=\"yellow\">A)</ansi> magnetism, <ansi fg=\"yellow\">B)</ansi> friction, <ansi fg=\"yellow\">C)</ansi> gravity",
        answers: ["c", "gravity"],
        step: "q2",
        nextStep: QUEST_ID + "-q3"
    },
    {
        question: "Water is made of hydrogen and what other element? <ansi fg=\"yellow\">A)</ansi> carbon, <ansi fg=\"yellow\">B)</ansi> helium, <ansi fg=\"yellow\">C)</ansi> oxygen",
        answers: ["c", "oxygen"],
        step: "q3",
        nextStep: QUEST_ID + "-end"
    }
];

function onAsk(mob, room, eventDetails) {
    if ( (user = GetUser(eventDetails.sourceId)) == null ) return false;

    quizMatch = UtilFindMatchIn(eventDetails.askText, quizSubjects);
    if ( !quizMatch.found ) return false;

    if ( !user.HasQuest(PREV_QUEST) ) {
        mob.Command("say Have you done Mr. Chen's math quiz yet? Science builds on math!");
        return true;
    }

    if ( user.HasQuest(QUEST_ID + "-end") ) {
        mob.Command("say You already passed! Try Ms. Flores in the Art Room next.");
        return true;
    }

    if ( !user.HasQuest(QUEST_ID + "-start") ) {
        mob.Command("emote snaps her goggles into place dramatically.");
        mob.Command("say Time for a science quiz! Three questions. Hypothesis: you'll do great.", 2);
        mob.Command("say <ansi fg=\"command\">Say</ansi> the letter or the answer. Let's experiment!", 4);
        mob.Command("say Question 1: " + quizQuestions[0].question, 6);
        user.GiveQuest(QUEST_ID + "-start");
        user.GiveQuest(QUEST_ID + "-q1");
        return true;
    }

    for ( i = 0; i < quizQuestions.length; i++ ) {
        if ( user.HasQuest(QUEST_ID + "-" + quizQuestions[i].step) && !user.HasQuest(quizQuestions[i].nextStep) ) {
            mob.Command("say " + quizQuestions[i].question);
            return true;
        }
    }
    return false;
}

function onSay(mob, room, eventDetails) {
    if ( (user = GetUser(eventDetails.sourceId)) == null ) return false;
    if ( !user.HasQuest(QUEST_ID + "-start") || user.HasQuest(QUEST_ID + "-end") ) return false;

    answer = eventDetails.msg.toLowerCase().trim();

    for ( i = 0; i < quizQuestions.length; i++ ) {
        q = quizQuestions[i];
        if ( !user.HasQuest(QUEST_ID + "-" + q.step) || user.HasQuest(q.nextStep) ) continue;

        correct = false;
        for ( a = 0; a < q.answers.length; a++ ) {
            if ( answer == q.answers[a] || answer.indexOf(q.answers[a]) >= 0 ) { correct = true; break; }
        }

        if ( correct ) {
            user.GiveQuest(q.nextStep);
            if ( q.nextStep == QUEST_ID + "-end" ) {
                mob.Command("say Correct! Hypothesis confirmed -- you're a natural scientist!");
                mob.Command("emote scribbles 'EXCELLENT' in her lab notebook.", 2);
            } else {
                nextIdx = i + 1;
                mob.Command("say Correct! The data supports your answer!");
                mob.Command("say Question " + String(nextIdx + 1) + ": " + quizQuestions[nextIdx].question, 2);
            }
            return true;
        } else {
            mob.Command("say Hmm, the results don't support that conclusion. Try again!");
            return true;
        }
    }
    return false;
}

function onIdle(mob, room) {
    if ( UtilGetRoundNumber() % 5 != 0 ) return false;

    missingQuizUsers = room.MissingQuest(QUEST_ID + "-start");
    if ( missingQuizUsers.length > 0 ) {
        mob.Command("say Pop quiz! <ansi fg=\"command\">Ask</ansi> me about the <ansi fg=\"command\">quiz</ansi>!");
        return true;
    }

    lines = [
        "emote peers into a microscope and says 'fascinating' to nobody.",
        "say Remember: correlation does not imply causation!",
        "emote feeds the class hamster a sunflower seed.",
    ];
    mob.Command(lines[UtilDiceRoll(1, lines.length) - 1]);
    return true;
}
