
// Console Room - Headquarters
// Maintains a permanent portal back to Riley's Bedroom (room 3)
// Also serves as the central hub with ambient flavor

function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, ["portal", "shimmer", "light", "exit", "back"]);
    if (!matches.found) {
        return false;
    }

    exits = room.GetExits();
    if (!exits["shimmering light"]) {
        room.AddTemporaryExit("shimmering light", ":rainbow", 3, "1 real day");
    }

    SendUserMessage(user.UserId(),
        'A ' + UtilApplyColorPattern('shimmering portal', 'rainbow') +
        ' pulses in the corner of the room, leading back to the real world. ' +
        'Type <ansi fg="command">shimmering light</ansi> to return to Riley\'s bedroom.');

    return true;
}

function onIdle(room) {

    // Ensure the portal back is always available
    exits = room.GetExits();
    if (!exits["shimmering light"]) {
        room.AddTemporaryExit("shimmering light", ":rainbow", 3, "1 real day");
    }

    return false;
}
