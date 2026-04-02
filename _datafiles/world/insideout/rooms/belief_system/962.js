
// Belief Forge - Quest 9 (The Belief Rewrite) interaction
// Player forges a new positive belief here

function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, ["forge", "anvil", "fire", "create", "belief", "craft"]);
    if (!matches.found) {
        return false;
    }

    if ((user = GetUser(user.UserId())) == null) {
        return false;
    }

    if (user.HasQuest("9-forge")) {
        SendUserMessage(user.UserId(),
            'The forge still glows warmly. The positive belief you created shimmers ' +
            'in your inventory, ready to be planted in the Sense of Self Core.');
        return true;
    }

    if (user.HasQuest("9-defeat")) {
        SendUserMessage(user.UserId(),
            'The Belief Forge burns with a steady, warm flame. Its surface is ' +
            'an anvil made of crystallized confidence. To forge a new belief, ' +
            'you need to combine a belief thread with a core memory.\n\n' +
            'Type <ansi fg="command">forge belief</ansi> to create a new positive belief.');
        return true;
    }

    SendUserMessage(user.UserId(),
        'The Belief Forge hums with potential. It could create something powerful, ' +
        'but you\'re not sure what you need yet.');
    return true;
}

function onCommand_forge(rest, user, room) {

    matches = UtilFindMatchIn(rest, ["belief", "thread", "positive"]);
    if (!matches.found) {
        return false;
    }

    if ((user = GetUser(user.UserId())) == null) {
        return false;
    }

    if (!user.HasQuest("9-defeat") || user.HasQuest("9-forge")) {
        return false;
    }

    SendUserMessage(user.UserId(),
        'You place the belief thread on the forge\'s anvil. The flame responds ' +
        'immediately, turning from orange to brilliant gold. The thread begins ' +
        'to glow, absorbing the warmth and light.\n\n' +
        'Words form along the thread as it\'s forged: "I AM GOOD ENOUGH."\n\n' +
        'The letters burn bright and true, each one a small act of defiance ' +
        'against every doubt and toxic thought that tried to tear Riley down. ' +
        'The finished belief floats up from the anvil, warm and radiant.\n\n' +
        'Take it to the Sense of Self Core to plant it where it belongs.');

    user.GiveQuest("9-forge");

    SendRoomMessage(room.RoomId(),
        'The Belief Forge blazes with golden light as a new belief is forged: ' +
        '"I AM GOOD ENOUGH."',
        user.UserId());

    return true;
}
