var quizSubjects = ["quiz", "test", "pop quiz", "exam", "math"];

var quizQuestions = [
    {
        question: "What is 7 times 8? <ansi fg=\"yellow\">A)</ansi> 54, <ansi fg=\"yellow\">B)</ansi> 56, <ansi fg=\"yellow\">C)</ansi> 58",
        answers: ["b", "56"],
        step: "q1",
        nextStep: "13-q2"
    },
    {
        question: "What is 3 + 4 x 2? <ansi fg=\"yellow\">A)</ansi> 14, <ansi fg=\"yellow\">B)</ansi> 10, <ansi fg=\"yellow\">C)</ansi> 11",
        answers: ["c", "11"],
        step: "q2",
        nextStep: "13-q3"
    },
    {
        question: "If x + 5 = 12, what is x? <ansi fg=\"yellow\">A)</ansi> 5, <ansi fg=\"yellow\">B)</ansi> 7, <ansi fg=\"yellow\">C)</ansi> 17",
        answers: ["b", "7"],
        step: "q3",
        nextStep: "13-end"
    }
];

function onAsk(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    quizMatch = UtilFindMatchIn(eventDetails.askText, quizSubjects);
    if ( !quizMatch.found ) {
        return false;
    }

    // Gate behind English quiz
    if ( !user.HasQuest("12-end") ) {
        mob.Command("say Have you taken Ms. Parker's English quiz yet? Start there first.");
        return true;
    }

    if ( user.HasQuest("13-end") ) {
        mob.Command("say You already passed! Elegant work. Try the Science Lab next -- Dr. Reeves has something for you.");
        return true;
    }

    if ( !user.HasQuest("13-start") ) {
        mob.Command("emote pushes his glasses up and grins.");
        mob.Command("say Ah, a willing participant! Three questions. Multiple choice.", 2);
        mob.Command("say Just <ansi fg=\"command\">say</ansi> the letter or the number. Here we go!", 4);
        mob.Command("say Question 1: " + quizQuestions[0].question, 6);
        user.GiveQuest("13-start");
        user.GiveQuest("13-q1");
        return true;
    }

    // Repeat current question
    for ( i = 0; i < quizQuestions.length; i++ ) {
        if ( user.HasQuest("13-" + quizQuestions[i].step) && !user.HasQuest(quizQuestions[i].nextStep) ) {
            mob.Command("say " + quizQuestions[i].question);
            return true;
        }
    }

    return false;
}

function onSay(mob, room, eventDetails) {

    if ( (user = GetUser(eventDetails.sourceId)) == null ) {
        return false;
    }

    if ( !user.HasQuest("13-start") || user.HasQuest("13-end") ) {
        return false;
    }

    answer = eventDetails.msg.toLowerCase().trim();

    for ( i = 0; i < quizQuestions.length; i++ ) {
        q = quizQuestions[i];

        if ( !user.HasQuest("13-" + q.step) || user.HasQuest(q.nextStep) ) {
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

            if ( q.nextStep == "13-end" ) {
                mob.Command("say Correct! Elegant! Absolutely elegant!");
                mob.Command("emote scribbles a gold star on his clipboard.", 2);
                mob.Command("say Perfect score. You've got a real head for numbers.", 4);
            } else {
                nextIdx = i + 1;
                mob.Command("say Correct! Beautiful!");
                mob.Command("say Question " + String(nextIdx + 1) + ": " + quizQuestions[nextIdx].question, 2);
            }
            return true;
        } else {
            mob.Command("say Not quite. The numbers don't add up! Try again.");
            return true;
        }
    }

    return false;
}

function onIdle(mob, room) {

    if ( UtilGetRoundNumber() % 5 != 0 ) {
        return false;
    }

    missingQuizUsers = room.MissingQuest("13-start");
    if ( missingQuizUsers.length > 0 ) {
        mob.Command("say Pop quiz today! <ansi fg=\"command\">Ask</ansi> me about the <ansi fg=\"command\">quiz</ansi> if you're brave enough!");
        return true;
    }

    lines = [
        "emote writes a complex equation on the board, then steps back to admire it.",
        "say The beauty of math is that there's always a right answer.",
        "emote taps the whiteboard. Elegant. Simply elegant.",
        "say Who can tell me what the square root of 144 is? Anyone?",
    ];

    mob.Command(lines[UtilDiceRoll(1, lines.length) - 1]);
    return true;
}
