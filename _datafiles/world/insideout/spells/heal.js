HEAL_DICE_QTY = 2;
HEAL_DICE_SIDES = 3;

function onCast(sourceActor, targetActor) {
    SendUserMessage(sourceActor.UserId(), 'You close your eyes and focus on a happy memory...');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' closes their eyes, concentrating.', sourceActor.UserId());
    return true;
}

function onWait(sourceActor, targetActor) {
    SendUserMessage(sourceActor.UserId(), 'A warm golden glow builds around your hands...');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+'\'s hands begin to glow warmly.', sourceActor.UserId());
}

function onMagic(sourceActor, targetActor) {
    roomId = sourceActor.GetRoomId();
    healAmt = UtilDiceRoll(HEAL_DICE_QTY, HEAL_DICE_SIDES);
    healAmtStr = String(healAmt);
    sourceUserId = sourceActor.UserId();
    sourceName = sourceActor.GetCharacterName(true);
    targetUserId = targetActor.UserId();
    targetName = targetActor.GetCharacterName(true);

    if (sourceActor.UserId() != targetActor.UserId()) {
        SendUserMessage(sourceUserId, 'You touch '+targetName+' with glowing hands, healing <ansi fg="healing">'+healAmtStr+' hitpoints</ansi>.');
        SendRoomMessage(roomId, sourceName+' touches '+targetName+' with glowing hands, providing comfort.', sourceUserId, targetUserId);
        SendUserMessage(targetUserId, sourceName+' touches you with warm, glowing hands, healing <ansi fg="healing">'+healAmtStr+' hitpoints</ansi>.');
    } else {
        SendUserMessage(sourceUserId, 'You embrace yourself with warm light, healing <ansi fg="healing">'+healAmtStr+' hitpoints</ansi>.');
        SendRoomMessage(roomId, sourceName+' is surrounded by a warm, healing glow.', sourceUserId, targetUserId);
    }
    targetActor.AddHealth(healAmt);
}
