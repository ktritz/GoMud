function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'A dreamy calm settles over your mind.');
}
function onTrigger(actor, triggersLeft) {
    manaAmt = actor.AddMana(UtilDiceRoll(1, 4));
    if (manaAmt > 0) {
        SendUserMessage(actor.UserId(), 'Dream energy restores <ansi fg="mana-100">'+String(manaAmt)+' mana</ansi>.');
    }
}
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'The dreamy focus fades away.');
}
