DMG_DICE_QTY = 1;
DMG_DICE_SIDES = 3;

function onCast(sourceActor, targetActors) {
    SendUserMessage(sourceActor.UserId(), 'You let your emotions surge, raw energy crackling around you...');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' crackles with emotional energy.', sourceActor.UserId());
    return true;
}

function onWait(sourceActor, targetActors) {
    SendUserMessage(sourceActor.UserId(), 'The emotional energy builds to a crescendo...');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' sparks with intensifying energy.', sourceActor.UserId());
}

function onMagic(sourceActor, targetActors) {
    roomId = sourceActor.GetRoomId();
    sourceUserId = sourceActor.UserId();
    sourceName = sourceActor.GetCharacterName(true);

    for (var i = 0; i < targetActors.length; i++) {
        dmgAmt = UtilDiceRoll(DMG_DICE_QTY, DMG_DICE_SIDES) + 1;
        dmgAmtStr = String(dmgAmt);
        targetUserId = targetActors[i].UserId();
        targetName = targetActors[i].GetCharacterName(true);

        if (sourceActor.UserId() != targetActors[i].UserId()) {
            SendUserMessage(sourceUserId, 'Emotional sparks strike '+targetName+' for <ansi fg="damage">'+dmgAmtStr+' damage</ansi>!');
            SendRoomMessage(roomId, sourceName+' unleashes emotional sparks that hit '+targetName+'!', sourceUserId, targetUserId);
            SendUserMessage(targetUserId, sourceName+' unleashes emotional sparks that hit you for <ansi fg="damage">'+dmgAmtStr+' damage</ansi>!');
        } else {
            SendUserMessage(sourceUserId, 'Your own emotional sparks zap you for <ansi fg="damage">'+dmgAmtStr+' damage</ansi>!');
            SendRoomMessage(roomId, sourceName+'\'s emotional sparks backfire!', sourceUserId, targetUserId);
        }
        targetActors[i].AddHealth(dmgAmt * -1);
    }
}
