function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'Reality begins to warp around you. You feel yourself... simplifying.');
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' begins to look... flatter.', actor.UserId());
}
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'Abstract thought continues to deconstruct you. Your edges are getting fuzzy.');
}
