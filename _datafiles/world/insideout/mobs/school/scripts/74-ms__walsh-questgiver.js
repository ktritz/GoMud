var quizSubjects = ["quiz", "test", "pop quiz", "exam", "geography", "map"];
var QUEST_ID = "17";
var PREV_QUEST = "16-end";

var quizQuestions = [
    {
        question: "What is the largest ocean on Earth? <ansi fg=\"yellow\">A)</ansi> Atlantic, <ansi fg=\"yellow\">B)</ansi> Pacific, <ansi fg=\"yellow\">C)</ansi> Indian",
        answers: ["b", "pacific"],
        step: "q1",
        nextStep: QUEST_ID + "-q2"
    },
    {
        question: "What continent is Egypt in? <ansi fg=\"yellow\">A)</ansi> Asia, <ansi fg=\"yellow\">B)</ansi> Europe, <ansi fg=\"yellow\">C)</ansi> Africa",
        answers: ["c", "africa"],
        step: "q2",
        nextStep: QUEST_ID + "-q3"
    },
    {
        question: "What is the capital of California? <ansi fg=\"yellow\">A)</ansi> Los Angeles, <ansi fg=\"yellow\">B)</ansi> Sacramento, <ansi fg=\"yellow\">C)</ansi> San Francisco",
        answers: ["b", "sacramento"],
        step: "q3",
        nextStep: QUEST_ID + "-end"
    }
];

function onAsk(mob, room, eventDetails) {
    if ( (user = GetUser(eventDetails.sourceId)) == null ) return false;
    quizMatch = UtilFindMatchIn(eventDetails.askText, quizSubjects);
    if ( !quizMatch.found ) return false;
    if ( !user.HasQuest(PREV_QUEST) ) { mob.Command("say Finish Mr. Ortiz's music quiz first!"); return true; }
    if ( user.HasQuest(QUEST_ID + "-end") ) { mob.Command("say You already passed! Try Mr. Grant's history quiz."); return true; }
    if ( !user.HasQuest(QUEST_ID + "-start") ) {
        mob.Command("emote spins the globe on her desk.");
        mob.Command("say Geography quiz! Three questions about our wonderful world.", 2);
        mob.Command("say <ansi fg=\"command\">Say</ansi> the letter or answer.", 4);
        mob.Command("say Question 1: " + quizQuestions[0].question, 6);
        user.GiveQuest(QUEST_ID + "-start"); user.GiveQuest(QUEST_ID + "-q1");
        return true;
    }
    for ( i = 0; i < quizQuestions.length; i++ ) {
        if ( user.HasQuest(QUEST_ID + "-" + quizQuestions[i].step) && !user.HasQuest(quizQuestions[i].nextStep) ) {
            mob.Command("say " + quizQuestions[i].question); return true;
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
        for ( a = 0; a < q.answers.length; a++ ) { if ( answer == q.answers[a] || answer.indexOf(q.answers[a]) >= 0 ) { correct = true; break; } }
        if ( correct ) {
            user.GiveQuest(q.nextStep);
            if ( q.nextStep == QUEST_ID + "-end" ) { mob.Command("say Correct! You really know your way around the world!"); }
            else { mob.Command("say Correct!"); mob.Command("say Question " + String(i+2) + ": " + quizQuestions[i+1].question, 2); }
            return true;
        } else { mob.Command("say Not quite. Check your atlas and try again!"); return true; }
    }
    return false;
}

function onIdle(mob, room) {
    if ( UtilGetRoundNumber() % 5 != 0 ) return false;
    if ( room.MissingQuest(QUEST_ID + "-start").length > 0 ) { mob.Command("say Geography quiz! <ansi fg=\"command\">Ask</ansi> me about the <ansi fg=\"command\">quiz</ansi>!"); return true; }
    lines = ["emote traces a route across the map with her finger.", "say Did you know there are 195 countries in the world?"];
    mob.Command(lines[UtilDiceRoll(1, lines.length) - 1]); return true;
}
