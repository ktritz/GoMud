HARM_DICE_QTY = 2;
HARM_DICE_SIDES = 6;
HARM_DICE_MOD = 1;

function onCast(sourceActor, targetActor) {
    SendUserMessage(sourceActor.UserId(), 'You reach deep into your memories, pulling at a core memory...');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' reaches out, a brilliant light forming.', sourceActor.UserId());
    return true;
}

function onWait(sourceActor, targetActor) {
    SendUserMessage(sourceActor.UserId(), 'The core memory blazes with power in your hands...');
    SendRoomMessage(sourceActor.GetRoomId(), 'A blazing orb of pure memory hovers before '+sourceActor.GetCharacterName(true)+'.', sourceActor.UserId());
}

function onMagic(sourceActor, targetActor) {
    roomId = sourceActor.GetRoomId();
    harmAmt = UtilDiceRoll(HARM_DICE_QTY, HARM_DICE_SIDES) + HARM_DICE_MOD;
    harmAmtStr = String(harmAmt);
    sourceUserId = sourceActor.UserId();
    sourceName = sourceActor.GetCharacterName(true);
    targetUserId = targetActor.UserId();
    targetName = targetActor.GetCharacterName(true);

    SendUserMessage(sourceUserId, 'You unleash the core memory blast at '+targetName+' for <ansi fg="damage">'+harmAmtStr+' damage</ansi>!');
    SendRoomMessage(roomId, sourceName+' hurls a blazing core memory at '+targetName+' with devastating force!', sourceUserId, targetUserId);
    SendUserMessage(targetUserId, sourceName+' hurls a blazing core memory at you for <ansi fg="damage">'+harmAmtStr+' damage</ansi>!');
    targetActor.AddHealth(harmAmt * -1);
}
