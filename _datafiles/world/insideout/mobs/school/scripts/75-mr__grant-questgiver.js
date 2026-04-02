var quizSubjects = ["quiz", "test", "pop quiz", "exam", "history"];
var QUEST_ID = "18";
var PREV_QUEST = "17-end";

var quizQuestions = [
    {
        question: "In what year did the United States declare independence? <ansi fg=\"yellow\">A)</ansi> 1776, <ansi fg=\"yellow\">B)</ansi> 1812, <ansi fg=\"yellow\">C)</ansi> 1492",
        answers: ["a", "1776"],
        step: "q1",
        nextStep: QUEST_ID + "-q2"
    },
    {
        question: "Who was the first president of the United States? <ansi fg=\"yellow\">A)</ansi> Thomas Jefferson, <ansi fg=\"yellow\">B)</ansi> Abraham Lincoln, <ansi fg=\"yellow\">C)</ansi> George Washington",
        answers: ["c", "george washington", "washington"],
        step: "q2",
        nextStep: QUEST_ID + "-q3"
    },
    {
        question: "What ancient civilization built the pyramids? <ansi fg=\"yellow\">A)</ansi> Romans, <ansi fg=\"yellow\">B)</ansi> Egyptians, <ansi fg=\"yellow\">C)</ansi> Greeks",
        answers: ["b", "egyptians", "egypt"],
        step: "q3",
        nextStep: QUEST_ID + "-end"
    }
];

function onAsk(mob, room, eventDetails) {
    if ( (user = GetUser(eventDetails.sourceId)) == null ) return false;
    quizMatch = UtilFindMatchIn(eventDetails.askText, quizSubjects);
    if ( !quizMatch.found ) return false;
    if ( !user.HasQuest(PREV_QUEST) ) { mob.Command("say Finish Ms. Walsh's geography quiz first!"); return true; }
    if ( user.HasQuest(QUEST_ID + "-end") ) { mob.Command("say You already passed! Try Coach Davis's PE quiz in the Gym."); return true; }
    if ( !user.HasQuest(QUEST_ID + "-start") ) {
        mob.Command("emote adjusts his bow tie with scholarly precision.");
        mob.Command("say A history quiz! Let's see if you know your past.", 2);
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
            if ( q.nextStep == QUEST_ID + "-end" ) { mob.Command("say Correct! History would be proud of you!"); }
            else { mob.Command("say Correct! Well done!"); mob.Command("say Question " + String(i+2) + ": " + quizQuestions[i+1].question, 2); }
            return true;
        } else { mob.Command("say That's not quite right. History remembers those who try again!"); return true; }
    }
    return false;
}

function onIdle(mob, room) {
    if ( UtilGetRoundNumber() % 5 != 0 ) return false;
    if ( room.MissingQuest(QUEST_ID + "-start").length > 0 ) { mob.Command("say History quiz! <ansi fg=\"command\">Ask</ansi> me about the <ansi fg=\"command\">quiz</ansi>!"); return true; }
    lines = ["emote flips through a history textbook with visible excitement.", "say History repeats itself. That's why I can give the same quiz every year."];
    mob.Command(lines[UtilDiceRoll(1, lines.length) - 1]); return true;
}
