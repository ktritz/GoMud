
// The Emotion Vault - Quest 8 (Anxiety's Overthrow) interaction
// Player frees the trapped emotions here

function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, ["vault", "emotions", "trapped", "free", "jars", "release"]);
    if (!matches.found) {
        return false;
    }

    if ((user = GetUser(user.UserId())) == null) {
        return false;
    }

    if (user.HasQuest("8-free")) {
        SendUserMessage(user.UserId(),
            'The vault is empty now. The jars are shattered, the emotions are free, ' +
            'and the room feels lighter than before. You should tell Joy the good news.');
        return true;
    }

    if (user.HasQuest("8-clear")) {
        SendUserMessage(user.UserId(),
            'The vault is lined with glass jars, each one containing a swirling, ' +
            'muted emotion. You can see faint colors through the glass -- golden joy, ' +
            'blue sadness, green disgust -- all dimmed and struggling. The emotions ' +
            'press against the glass, silently begging to be released.\n\n' +
            'A control panel on the wall has a single large button labeled ' +
            '"EMERGENCY RELEASE." It looks like it was installed by someone who was ' +
            'very anxious about needing an exit strategy.\n\n' +
            'Type <ansi fg="command">push button</ansi> to free the emotions.');
        return true;
    }

    SendUserMessage(user.UserId(),
        'Glass jars line the walls, some still faintly glowing with trapped emotional ' +
        'energy. This is where Anxiety locked away the emotions she thought were ' +
        'holding Riley back.');
    return true;
}

function onCommand_push(rest, user, room) {

    matches = UtilFindMatchIn(rest, ["button", "release", "emergency"]);
    if (!matches.found) {
        return false;
    }

    if ((user = GetUser(user.UserId())) == null) {
        return false;
    }

    if (!user.HasQuest("8-clear") || user.HasQuest("8-free")) {
        return false;
    }

    SendUserMessage(user.UserId(),
        'You slam the EMERGENCY RELEASE button. Every jar in the vault cracks ' +
        'simultaneously, and brilliant color EXPLODES outward -- golden light, ' +
        'deep blue waves, fiery red, vivid purple, sharp green. The emotions ' +
        'pour out of their containers and swirl together in a dazzling tornado ' +
        'of feeling.\n\n' +
        'For a moment, the room is pure chaos. Then the colors settle, separate, ' +
        'and stream upward through the ceiling, heading back to Headquarters.\n\n' +
        'A quiet voice says, "...thank you." You\'re not sure which emotion said it. ' +
        'Maybe all of them.');

    user.GiveQuest("8-free");

    SendRoomMessage(room.RoomId(),
        'An explosion of color bursts from the vault as trapped emotions are freed! ' +
        'Brilliant lights stream upward through the ceiling.',
        user.UserId());

    return true;
}
