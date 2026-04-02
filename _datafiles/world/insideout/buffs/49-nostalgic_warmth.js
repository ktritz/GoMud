function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'A wave of bittersweet nostalgia washes over you...');
}
function onTrigger(actor, triggersLeft) {
    healAmt = actor.AddHealth(UtilDiceRoll(1, 6));
    if (healAmt > 0) {
        SendUserMessage(actor.UserId(), 'Warm memories heal you for <ansi fg="healing">'+String(healAmt)+' health</ansi>.');
    }
}
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'The nostalgic feeling fades, leaving a gentle warmth behind.');
}
