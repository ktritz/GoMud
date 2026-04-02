function onCast(sourceActor, targetActor) {
    SendUserMessage(sourceActor.UserId(), 'You focus on clearing away the toxic thoughts...');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' focuses intently, a clean light building.', sourceActor.UserId());
    return true;
}

function onWait(sourceActor, targetActor) {
    SendUserMessage(sourceActor.UserId(), 'The purifying light grows stronger...');
    SendRoomMessage(sourceActor.GetRoomId(), sourceActor.GetCharacterName(true)+' glows with purifying light.', sourceActor.UserId());
}

function onMagic(sourceActor, targetActor) {
    roomId = sourceActor.GetRoomId();
    sourceUserId = sourceActor.UserId();
    sourceName = sourceActor.GetCharacterName(true);
    targetUserId = targetActor.UserId();
    targetName = targetActor.GetCharacterName(true);

    if (sourceActor.UserId() != targetActor.UserId()) {
        SendUserMessage(sourceUserId, 'You purify '+targetName+'\'s thoughts, cleansing away the toxins.');
        SendRoomMessage(roomId, sourceName+' directs purifying light towards '+targetName+'.', sourceUserId, targetUserId);
        SendUserMessage(targetUserId, sourceName+' purifies your thoughts. The toxins fade away.');
    } else {
        SendUserMessage(sourceUserId, 'You purify your own thoughts. The toxins dissolve.');
        SendRoomMessage(roomId, sourceName+' glows briefly as toxic thoughts are purged.', sourceUserId);
    }
    targetActor.CancelBuffWithFlag("poison");
}
