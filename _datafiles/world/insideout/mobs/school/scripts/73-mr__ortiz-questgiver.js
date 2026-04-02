var quizSubjects = ["quiz", "test", "pop quiz", "exam", "music"];
var QUEST_ID = "16";
var PREV_QUEST = "15-end"; // Must complete art first

var quizQuestions = [
    {
        question: "How many notes are in a musical octave? <ansi fg=\"yellow\">A)</ansi> 6, <ansi fg=\"yellow\">B)</ansi> 8, <ansi fg=\"yellow\">C)</ansi> 12",
        answers: ["b", "8", "eight"],
        step: "q1",
        nextStep: QUEST_ID + "-q2"
    },
    {
        question: "What does 'forte' mean in music? <ansi fg=\"yellow\">A)</ansi> play slowly, <ansi fg=\"yellow\">B)</ansi> play softly, <ansi fg=\"yellow\">C)</ansi> play loudly",
        answers: ["c", "loudly", "play loudly", "loud"],
        step: "q2",
        nextStep: QUEST_ID + "-q3"
    },
    {
        question: "Which instrument has 88 keys? <ansi fg=\"yellow\">A)</ansi> guitar, <ansi fg=\"yellow\">B)</ansi> piano, <ansi fg=\"yellow\">C)</ansi> violin",
        answers: ["b", "piano"],
        step: "q3",
        nextStep: QUEST_ID + "-end"
    }
];

function onAsk(mob, room, eventDetails) {
    if ( (user = GetUser(eventDetails.sourceId)) == null ) return false;

    quizMatch = UtilFindMatchIn(eventDetails.askText, quizSubjects);
    if ( !quizMatch.found ) return false;

    if ( !user.HasQuest(PREV_QUEST) ) {
        mob.Command("say Music is the final frontier! But finish Ms. Flores' art quiz first.");
        return true;
    }

    if ( user.HasQuest(QUEST_ID + "-end") ) {
        mob.Command("say You already passed! You've completed all the quizzes. The principal might have something special for you.");
        return true;
    }

    if ( !user.HasQuest(QUEST_ID + "-start") ) {
        mob.Command("emote spins a drumstick between his fingers.");
        mob.Command("say A music quiz! Let's see if you've got rhythm AND knowledge.", 2);
        mob.Command("say <ansi fg=\"command\">Say</ansi> the letter or the answer. And a-one, and a-two...", 4);
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
                mob.Command("say Correct! Standing ovation! You know your music!");
                mob.Command("emote plays a triumphant riff on air guitar.", 2);
                mob.Command("say You've completed all the pop quizzes! Talk to the principal for something special.", 4);
            } else {
                nextIdx = i + 1;
                mob.Command("say Correct! That's music to my ears!");
                mob.Command("say Question " + String(nextIdx + 1) + ": " + quizQuestions[nextIdx].question, 2);
            }
            return true;
        } else {
            mob.Command("say That's a little off-key. Try again!");
            return true;
        }
    }
    return false;
}

function onIdle(mob, room) {
    if ( UtilGetRoundNumber() % 5 != 0 ) return false;

    missingQuizUsers = room.MissingQuest(QUEST_ID + "-start");
    if ( missingQuizUsers.length > 0 ) {
        mob.Command("say Got music in your soul? <ansi fg=\"command\">Ask</ansi> me about the <ansi fg=\"command\">quiz</ansi>!");
        return true;
    }

    lines = [
        "emote drums a complex rhythm on the desk with two pencils.",
        "say If you can hum it, you can play it. Probably.",
        "emote strums a guitar quietly in the corner.",
    ];
    mob.Command(lines[UtilDiceRoll(1, lines.length) - 1]);
    return true;
}
