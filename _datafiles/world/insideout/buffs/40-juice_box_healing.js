function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'You take a sip of the juice box. Refreshing!');
}
function onTrigger(actor, triggersLeft) {
    healAmt = actor.AddHealth(UtilDiceRoll(1, 4));
    if (healAmt > 0) {
        SendUserMessage(actor.UserId(), 'The juice continues to refresh you for <ansi fg="healing">'+String(healAmt)+' health</ansi>.');
    }
}
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'You finish the last drops of juice.');
}
