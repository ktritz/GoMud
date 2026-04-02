
// Riley's Bedroom - Portal to the Mindscape
// When a player looks at the bed/shimmer/light, a temporary exit opens to Headquarters (room 200)

function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, ["bed", "under", "shimmer", "light", "glow", "under bed"]);
    if (!matches.found) {
        return false;
    }

    SendUserMessage(user.UserId(),
        'You peer under Riley\'s bed and notice a strange ' +
        UtilApplyColorPattern('shimmering light', 'rainbow') +
        ' pulsing softly beneath the frame, like a heartbeat made of color. ' +
        'It seems to be calling to you...');

    exits = room.GetExits();
    if (!exits["shimmering light"]) {
        room.AddTemporaryExit("shimmering light", ":rainbow", 200, "1 real day");
        SendRoomMessage(room.RoomId(),
            'A ' + UtilApplyColorPattern('shimmering portal', 'rainbow') +
            ' opens beneath the bed, revealing a passage into Riley\'s mind!',
            user.UserId());
    }

    SendUserMessage(user.UserId(),
        '\nType <ansi fg="command">shimmering light</ansi> to enter Riley\'s mind.\n' +
        '<ansi fg="yellow">Warning: The mindscape can be dangerous. Headquarters is safe, but ' +
        'venturing beyond it into Long Term Memory and other zones is not recommended below level 4. ' +
        'Explore the neighborhood and school first!</ansi>');

    return true;
}
