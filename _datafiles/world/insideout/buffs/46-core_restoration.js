function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'The core memory essence pulses through you with golden warmth.');
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' glows with a golden light.', actor.UserId());
}
function onTrigger(actor, triggersLeft) {
    healAmt = actor.AddHealth(UtilDiceRoll(2, 6));
    if (healAmt > 0) {
        SendUserMessage(actor.UserId(), 'Core memory energy restores <ansi fg="healing">'+String(healAmt)+' health</ansi>.');
    }
}
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'The golden glow fades as the core memory energy is spent.');
}
