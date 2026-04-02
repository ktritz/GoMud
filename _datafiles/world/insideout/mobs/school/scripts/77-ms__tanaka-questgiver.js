var quizSubjects = ["quiz", "test", "pop quiz", "exam", "programming", "code", "coding", "computer"];
var QUEST_ID = "20";
var PREV_QUEST = "19-end";

var quizQuestions = [
    {
        question: "What does HTML stand for? <ansi fg=\"yellow\">A)</ansi> How To Make Links, <ansi fg=\"yellow\">B)</ansi> HyperText Markup Language, <ansi fg=\"yellow\">C)</ansi> Home Tool Markup Language",
        answers: ["b", "hypertext markup language", "hypertext"],
        step: "q1",
        nextStep: QUEST_ID + "-q2"
    },
    {
        question: "In programming, what is a 'bug'? <ansi fg=\"yellow\">A)</ansi> a type of virus, <ansi fg=\"yellow\">B)</ansi> an error in code, <ansi fg=\"yellow\">C)</ansi> a feature request",
        answers: ["b", "error", "an error"],
        step: "q2",
        nextStep: QUEST_ID + "-q3"
    },
    {
        question: "What symbol starts a comment in Python? <ansi fg=\"yellow\">A)</ansi> //, <ansi fg=\"yellow\">B)</ansi> #, <ansi fg=\"yellow\">C)</ansi> --",
        answers: ["b", "#", "hash"],
        step: "q3",
        nextStep: QUEST_ID + "-end"
    }
];

function onAsk(mob, room, eventDetails) {
    if ( (user = GetUser(eventDetails.sourceId)) == null ) return false;
    quizMatch = UtilFindMatchIn(eventDetails.askText, quizSubjects);
    if ( !quizMatch.found ) return false;
    if ( !user.HasQuest(PREV_QUEST) ) { mob.Command("say Finish Coach Davis's PE quiz first. Healthy body, healthy coder."); return true; }
    if ( user.HasQuest(QUEST_ID + "-end") ) { mob.Command("say You already passed! You've completed all the school quizzes. Talk to the principal for something special."); return true; }
    if ( !user.HasQuest(QUEST_ID + "-start") ) {
        mob.Command("emote cracks her knuckles over the keyboard.");
        mob.Command("say Programming quiz. Three questions. No stack overflow allowed.", 2);
        mob.Command("say <ansi fg=\"command\">Say</ansi> the letter or answer. Compile your thoughts.", 4);
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
            if ( q.nextStep == QUEST_ID + "-end" ) {
                mob.Command("say Correct! Code compiled successfully. Zero errors.");
                mob.Command("emote gives a rare, approving nod.", 2);
                mob.Command("say You've completed all the school quizzes. The principal might have something for you.", 4);
            }
            else { mob.Command("say Correct! Clean code."); mob.Command("say Question " + String(i+2) + ": " + quizQuestions[i+1].question, 2); }
            return true;
        } else { mob.Command("say Syntax error. Try again."); return true; }
    }
    return false;
}

function onIdle(mob, room) {
    if ( UtilGetRoundNumber() % 5 != 0 ) return false;
    if ( room.MissingQuest(QUEST_ID + "-start").length > 0 ) { mob.Command("say Coding quiz! <ansi fg=\"command\">Ask</ansi> me about the <ansi fg=\"command\">quiz</ansi>!"); return true; }
    lines = ["emote types rapidly, pauses, then deletes everything.", "say Debugging is twice as hard as writing code. So if you write code as cleverly as possible, you're not smart enough to debug it."];
    mob.Command(lines[UtilDiceRoll(1, lines.length) - 1]); return true;
}
