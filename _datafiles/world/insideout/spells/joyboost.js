HEAL_DICE_QTY = 2;
HEAL_DICE_SIDES = 4;
HEAL_DICE_MOD = 2;

function onCast(sourceActor, targetActor) {
    SendUserMessage(sourceActor.UserId(), 'You think of something that makes you truly happy...');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' smiles warmly, golden light building around them.', sourceActor.UserId());
    return true;
}

function onWait(sourceActor, targetActor) {
    SendUserMessage(sourceActor.UserId(), 'Joy fills you up from the inside...');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' radiates pure, golden joy.', sourceActor.UserId());
}

function onMagic(sourceActor, targetActor) {
    roomId = sourceActor.GetRoomId();
    healAmt = UtilDiceRoll(HEAL_DICE_QTY, HEAL_DICE_SIDES) + HEAL_DICE_MOD;
    healAmtStr = String(healAmt);
    sourceUserId = sourceActor.UserId();
    sourceName = sourceActor.GetCharacterName(true);
    targetUserId = targetActor.UserId();
    targetName = targetActor.GetCharacterName(true);

    if (sourceActor.UserId() != targetActor.UserId()) {
        SendUserMessage(sourceUserId, 'Joy flows from you into '+targetName+', healing <ansi fg="healing">'+healAmtStr+' hitpoints</ansi>!');
        SendRoomMessage(roomId, 'Golden light flows from '+sourceName+' into '+targetName+'.', sourceUserId, targetUserId);
        SendUserMessage(targetUserId, sourceName+' fills you with joy, healing <ansi fg="healing">'+healAmtStr+' hitpoints</ansi>!');
    } else {
        SendUserMessage(sourceUserId, 'Joy fills every part of you, healing <ansi fg="healing">'+healAmtStr+' hitpoints</ansi>!');
        SendRoomMessage(roomId, sourceName+' glows with radiant golden joy.', sourceUserId, targetUserId);
    }
    targetActor.AddHealth(healAmt);
}
