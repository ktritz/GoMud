
// Balcony - warns low-level players before descending to Long Term Memory

function onCommand_stairway(rest, user, room) {

    level = user.GetLevel();
    if (level < 4) {
        SendUserMessage(user.UserId(),
            '<ansi fg="yellow">You peer down the stairway into the vast shelves of Long Term Memory below. ' +
            'Something moves in the shadows between the memory orbs. ' +
            'This doesn\'t look safe for someone your level.</ansi>\n' +
            '<ansi fg="yellow">Recommended level: 4+. Type <ansi fg="command">stairway</ansi> again to descend anyway.</ansi>');

        warnCount = user.GetMiscCharacterData('ltm_stair_warn');
        if (warnCount === null || warnCount < 1) {
            user.SetMiscCharacterData('ltm_stair_warn', 1);
            return true;
        }
        user.SetMiscCharacterData('ltm_stair_warn', 0);
    }
    return false;
}
