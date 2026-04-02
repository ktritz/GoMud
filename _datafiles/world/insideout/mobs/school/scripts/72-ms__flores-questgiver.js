var quizSubjects = ["quiz", "test", "pop quiz", "exam", "art"];
var QUEST_ID = "15";
var PREV_QUEST = "14-end"; // Must complete science first

var quizQuestions = [
    {
        question: "What are the three primary colors? <ansi fg=\"yellow\">A)</ansi> red green blue, <ansi fg=\"yellow\">B)</ansi> red yellow blue, <ansi fg=\"yellow\">C)</ansi> red orange purple",
        answers: ["b", "red yellow blue"],
        step: "q1",
        nextStep: QUEST_ID + "-q2"
    },
    {
        question: "What artist painted the Mona Lisa? <ansi fg=\"yellow\">A)</ansi> Picasso, <ansi fg=\"yellow\">B)</ansi> Van Gogh, <ansi fg=\"yellow\">C)</ansi> Da Vinci",
        answers: ["c", "da vinci", "leonardo"],
        step: "q2",
        nextStep: QUEST_ID + "-q3"
    },
    {
        question: "What is a self-portrait? <ansi fg=\"yellow\">A)</ansi> a painting of a landscape, <ansi fg=\"yellow\">B)</ansi> a painting of yourself, <ansi fg=\"yellow\">C)</ansi> a painting of food",
        answers: ["b", "yourself", "painting of yourself"],
        step: "q3",
        nextStep: QUEST_ID + "-end"
    }
];

function onAsk(mob, room, eventDetails) {
    if ( (user = GetUser(eventDetails.sourceId)) == null ) return false;

    quizMatch = UtilFindMatchIn(eventDetails.askText, quizSubjects);
    if ( !quizMatch.found ) return false;

    if ( !user.HasQuest(PREV_QUEST) ) {
        mob.Command("say Art connects to everything! But try Dr. Reeves' science quiz first.");
        return true;
    }

    if ( user.HasQuest(QUEST_ID + "-end") ) {
        mob.Command("say You already passed, you creative genius! Try Mr. Ortiz in the Music Room.");
        return true;
    }

    if ( !user.HasQuest(QUEST_ID + "-start") ) {
        mob.Command("emote claps her paint-stained hands together.");
        mob.Command("say An art quiz! How delightful! Three questions about creativity and art history.", 2);
        mob.Command("say <ansi fg=\"command\">Say</ansi> the letter or the answer. Let your knowledge flow!", 4);
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
                mob.Command("say Correct! Oh, that's beautiful! You have a real eye for this!");
                mob.Command("emote sketches a tiny star on your hand in marker.", 2);
            } else {
                nextIdx = i + 1;
                mob.Command("say Correct! Wonderful!");
                mob.Command("say Question " + String(nextIdx + 1) + ": " + quizQuestions[nextIdx].question, 2);
            }
            return true;
        } else {
            mob.Command("say Hmm, not quite. Art is about seeing things differently -- try again!");
            return true;
        }
    }
    return false;
}

function onIdle(mob, room) {
    if ( UtilGetRoundNumber() % 5 != 0 ) return false;

    missingQuizUsers = room.MissingQuest(QUEST_ID + "-start");
    if ( missingQuizUsers.length > 0 ) {
        mob.Command("say Feeling creative? <ansi fg=\"command\">Ask</ansi> me about the <ansi fg=\"command\">quiz</ansi>!");
        return true;
    }

    lines = [
        "emote mixes two paint colors together and gasps at the result.",
        "say Art is not what you see, but what you make others see.",
        "emote hums while arranging dried flowers into something that might be a sculpture.",
    ];
    mob.Command(lines[UtilDiceRoll(1, lines.length) - 1]);
    return true;
}
