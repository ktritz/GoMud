
// The Dot - Quest 7 (Abstract Rescue) interaction
// Player finds the deconstructed Mind Worker here and can reconstruct them

function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, ["dot", "worker", "mind", "reconstruct", "help", "point"]);
    if (!matches.found) {
        return false;
    }

    if ((user = GetUser(user.UserId())) == null) {
        return false;
    }

    if (user.HasQuest("7-rescue")) {
        SendUserMessage(user.UserId(),
            'The Mind Worker is fully reconstructed and looking shaky but grateful. ' +
            'You should escort them back to safety in Long Term Memory.');
        return true;
    }

    if (user.HasQuest("7-find")) {
        SendUserMessage(user.UserId(),
            'At the very center of the room, you see it -- a single, vibrating dot. ' +
            'It\'s the Mind Worker, reduced to their most abstract form. A tiny, ' +
            'pulsing point of existence that somehow still manages to look terrified.\n\n' +
            'If you have an abstract thought fragment, you might be able to use it ' +
            'to help rebuild them. Try <ansi fg="command">give fragment</ansi>.');
        return true;
    }

    SendUserMessage(user.UserId(),
        'A single point floats at the center of the room. It vibrates with ' +
        'a faint urgency, as if it\'s trying to be more than it currently is.');
    return true;
}

function onCommand_give(rest, user, room) {

    matches = UtilFindMatchIn(rest, ["fragment", "abstract", "thought"]);
    if (!matches.found) {
        return false;
    }

    if ((user = GetUser(user.UserId())) == null) {
        return false;
    }

    if (!user.HasQuest("7-find")) {
        return false;
    }

    if (user.HasQuest("7-rescue")) {
        return false;
    }

    SendUserMessage(user.UserId(),
        'You hold out the abstract thought fragment toward the dot. The fragment ' +
        'resonates, vibrates, and then SHATTERS into a cascade of geometric light. ' +
        'The dot begins to expand -- a line, then a shape, then a face, then a body. ' +
        'The Mind Worker rebuilds before your eyes, clipboard and all.\n\n' +
        '"Oh thank goodness!" the worker gasps, patting themselves down. "I still have ' +
        'my arms! And my clipboard! I thought I was going to be a dot FOREVER!" ' +
        'They look at you with enormous gratitude. "Please, get me out of here!"');

    user.GiveQuest("7-rescue");

    SendRoomMessage(room.RoomId(),
        'A cascade of geometric light fills the room as a Mind Worker is ' +
        'reconstructed from a single point!',
        user.UserId());

    return true;
}
