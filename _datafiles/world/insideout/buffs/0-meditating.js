function onStart(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'You sit down and close your eyes, letting your thoughts settle.');
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' sits down and begins to meditate.', actor.UserId());
}
function onTrigger(actor, triggersLeft) {
    SendUserMessage(actor.UserId(), 'Your meditation continues. <ansi bg="blue"> *' + triggersLeft + ' rounds left* </ansi>');
    SendRoomMessage(actor.GetRoomId(), actor.GetCharacterName(true)+' continues meditating peacefully.', actor.UserId());
}
