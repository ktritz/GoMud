function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'A soft, warm glow surrounds you, pushing back the darkness.');
    SendRoomMessage(actor.GetRoomId(), 'A warm glow surrounds '+actor.GetCharacterName(true)+'.', actor.UserId());
}
function onEnd(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'Your glow slowly fades away.');
    SendRoomMessage(actor.GetRoomId(), 'The glow surrounding '+actor.GetCharacterName(true)+' fades away.', actor.UserId());
}
