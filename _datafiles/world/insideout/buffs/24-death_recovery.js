function onTrigger(actor, triggersLeft) {
    healAmt = actor.AddHealth(UtilDiceRoll(1, 10));
    manaAmt = actor.AddMana(UtilDiceRoll(1, 10));
    if (healAmt > 0 && manaAmt > 0) {
        SendUserMessage(actor.UserId(), 'The mindscape slowly restores you: <ansi fg="healing">'+String(healAmt)+' health</ansi> and <ansi fg="mana-100">'+String(manaAmt)+' mana</ansi>.');
    } else if (healAmt > 0) {
        SendUserMessage(actor.UserId(), 'The mindscape restores <ansi fg="healing">'+String(healAmt)+' health</ansi>.');
    } else if (manaAmt > 0) {
        SendUserMessage(actor.UserId(), 'The mindscape restores <ansi fg="mana-100">'+String(manaAmt)+' mana</ansi>.');
    }
    if (healAmt > 0 || manaAmt > 0) {
        SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' is slowly recovering.', actor.UserId());
    }
}
