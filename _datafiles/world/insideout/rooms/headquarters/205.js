
// Recall Tube - warns low-level players before entering Long Term Memory

function onCommand_tube(rest, user, room) {

    level = user.GetLevel();
    if (level < 4) {
        SendUserMessage(user.UserId(),
            '<ansi fg="yellow">The Mind Worker at the tube station gives you a worried look.</ansi>\n' +
            '"Hey kid, you sure about this? Long Term Memory can be rough -- ' +
            'there are Memory Leeches and Gloom Raiders down there. ' +
            'I wouldn\'t go below level 4 if I were you."\n' +
            '<ansi fg="yellow">Type <ansi fg="command">tube</ansi> again if you\'re sure.</ansi>');

        warnCount = user.GetMiscCharacterData('ltm_warn');
        if (warnCount === null || warnCount < 1) {
            user.SetMiscCharacterData('ltm_warn', 1);
            return true; // block the first attempt
        }
        // Let them through on second try
        user.SetMiscCharacterData('ltm_warn', 0);
    }
    return false;
}
