HARM_DICE_QTY = 1;
HARM_DICE_SIDES = 6;
HARM_DICE_MOD = 2;

function onCast(sourceActor, targetActor) {
    SendUserMessage(sourceActor.UserId(), 'You concentrate, pulling fragments of memory together...');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' concentrates intensely, light gathering around them.', sourceActor.UserId());
    return true;
}

function onWait(sourceActor, targetActor) {
    SendUserMessage(sourceActor.UserId(), 'A shimmering memory orb forms in your hand...');
    SendRoomMessage(sourceActor.GetRoomId(), 'A glowing orb forms in '+sourceActor.GetCharacterName(true)+'\'s hand.', sourceActor.UserId());
}

function onMagic(sourceActor, targetActor) {
    roomId = sourceActor.GetRoomId();
    harmAmt = UtilDiceRoll(HARM_DICE_QTY, HARM_DICE_SIDES) + HARM_DICE_MOD;
    harmAmtStr = String(harmAmt);
    sourceUserId = sourceActor.UserId();
    sourceName = sourceActor.GetCharacterName(true);
    targetUserId = targetActor.UserId();
    targetName = targetActor.GetCharacterName(true);

    SendUserMessage(sourceUserId, 'You hurl the memory missile at '+targetName+', doing <ansi fg="damage">'+harmAmtStr+' damage</ansi>!');
    SendRoomMessage(roomId, sourceName+' hurls a glowing memory orb at '+targetName+'!', sourceUserId, targetUserId);
    SendUserMessage(targetUserId, sourceName+' hurls a glowing memory orb at you for <ansi fg="damage">'+harmAmtStr+' damage</ansi>!');
    targetActor.AddHealth(harmAmt * -1);
}
