
// Emotional Tag Station - Quest 5 (Sadness Has a Point) interaction
// When a player looks at or interacts with the equipment, they learn about memory recoloring

function onCommand_look(rest, user, room) {

    matches = UtilFindMatchIn(rest, ["equipment", "station", "tag", "recolor", "machine", "memories"]);
    if (!matches.found) {
        return false;
    }

    if ((user = GetUser(user.UserId())) == null) {
        return false;
    }

    if (user.HasQuest("5-tag")) {
        SendUserMessage(user.UserId(),
            'You\'ve already seen how the Tag Station works. The memory you recolored ' +
            'still sits on the conveyor, shifting gently between blue and gold. ' +
            'Maybe you should tell Sadness what you learned.');
        return true;
    }

    if (user.HasQuest("5-start")) {
        SendUserMessage(user.UserId(),
            'You examine the Emotional Tag Station closely. A memory orb sits in the ' +
            'recoloring cradle -- it\'s a golden joy memory of Riley laughing with friends. ' +
            'As you watch, a Mind Worker adjusts the settings and the orb slowly shifts ' +
            'from pure gold to a bittersweet blue-gold swirl.\n\n' +
            'The worker notices you watching. "See that? Happy memories aren\'t always ' +
            'just happy. Sometimes the sad parts are what make them meaningful. That ' +
            'laughing-with-friends memory? It\'s also a missing-those-friends memory now ' +
            'that Riley moved. Both feelings are true at the same time."\n\n' +
            'You realize something important: sadness isn\'t the opposite of joy. ' +
            'It\'s what gives joy its depth.');

        user.GiveQuest("5-tag");

        SendRoomMessage(room.RoomId(),
            user.GetCharacterName(true) + ' watches a memory being recolored at the Tag Station, ' +
            'and seems to understand something important.',
            user.UserId());
        return true;
    }

    SendUserMessage(user.UserId(),
        'The Emotional Tag Station hums quietly. Memory orbs pass through on a conveyor, ' +
        'and workers carefully adjust their emotional coloring. It\'s fascinating to watch, ' +
        'but you\'re not sure why you\'d need to use it right now.');
    return true;
}
