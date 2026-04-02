
// Sense of Self Core - Quest 9 (The Belief Rewrite) - plant the new belief

function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, ["core", "center", "self", "plant", "belief"]);
    if (!matches.found) {
        return false;
    }

    if ((user = GetUser(user.UserId())) == null) {
        return false;
    }

    if (user.HasQuest("9-plant")) {
        SendUserMessage(user.UserId(),
            'The Sense of Self Core pulses with warm golden light. The new belief ' +
            '"I AM GOOD ENOUGH" is woven into its very fabric, glowing steadily.');
        return true;
    }

    if (user.HasQuest("9-forge")) {
        SendUserMessage(user.UserId(),
            'The Sense of Self Core floats at the center of the room -- a brilliant, ' +
            'slowly rotating sphere of interwoven beliefs. You can see threads of ' +
            '"I am kind" and "I am brave" woven through it, but there\'s a dark ' +
            'spot where "I\'m not good enough" left a hole.\n\n' +
            'The new belief you forged would fit perfectly there.\n\n' +
            'Type <ansi fg="command">plant belief</ansi> to weave it into Riley\'s sense of self.');
        return true;
    }

    SendUserMessage(user.UserId(),
        'The Sense of Self Core is a beautiful, complex sphere of woven beliefs. ' +
        'It represents everything Riley believes about herself.');
    return true;
}

function onCommand_plant(rest, user, room) {

    matches = UtilFindMatchIn(rest, ["belief", "thread"]);
    if (!matches.found) {
        return false;
    }

    if ((user = GetUser(user.UserId())) == null) {
        return false;
    }

    if (!user.HasQuest("9-forge") || user.HasQuest("9-plant")) {
        return false;
    }

    SendUserMessage(user.UserId(),
        'You hold the glowing belief thread up to the Sense of Self Core. ' +
        'The sphere reaches out to it, golden threads extending like welcoming arms. ' +
        'The new belief -- "I AM GOOD ENOUGH" -- weaves itself into the core, ' +
        'filling the dark gap left by the toxic thought.\n\n' +
        'The entire sphere BLAZES with light. Colors ripple outward -- gold, blue, ' +
        'red, purple, green -- and for a moment you feel what Riley feels: ' +
        'a deep, warm certainty that she is exactly who she\'s supposed to be.\n\n' +
        'The toxic belief is gone. In its place, something stronger. Something true.');

    user.GiveQuest("9-plant");

    SendRoomMessage(room.RoomId(),
        'The Sense of Self Core blazes with renewed light as a new belief takes root. ' +
        'A wave of warmth and confidence radiates outward through the entire Belief System.',
        user.UserId());

    return true;
}
