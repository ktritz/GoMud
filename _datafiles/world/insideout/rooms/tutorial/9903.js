
const allowed_commands = ["help", "broadcast", "look", "status", "inventory", "experience", "conditions", "equip"];
const teach_commands = ["get cap", "equip cap", "portal"];
const teacherMobId = 97;
const teacherName = "Orb of Adventure";
const capItemId = 20001;

var commandNow = 0; // Which command they are on



// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    ignoreCommand = false;

    teacherMob = getTeacher(room);

    var extraDelay = 0;

    fullCommand = cmd;
    if ( rest.length > 0 ) {
        fullCommand = cmd + ' ' + rest;
    }

    if ( commandNow >= 2 ) {
        return false;
    }

    if ( teach_commands[commandNow] == fullCommand ) {

        if ( fullCommand == "equip cap" ) {
            teacherMob.Command("say Good job!", 1.0);
        } else {
            teacherMob.Command("say Good job! You earned it!", 1.0);
        }

        extraDelay = 1.0;

        commandNow++;

    } else {

        if ( allowed_commands.includes(cmd) || teach_commands.slice(0, commandNow).includes(cmd) ) {
            return false;
        }

        ignoreCommand = true;
    }

    switch (commandNow) {
        case 0:
            teacherMob.Command('emote gestures to the <ansi fg="item">baseball cap</ansi> on the ground.', extraDelay+2.0);
            teacherMob.Command('say type <ansi fg="command">get cap</ansi> to pick up the <ansi fg="item">baseball cap</ansi>.', extraDelay+3.0);
            break;
        case 1:
            teacherMob.Command('say Go ahead and wear the <ansi fg="item">baseball cap</ansi> by typing <ansi fg="command">equip cap</ansi>.', extraDelay+2.0);
            break;
        case 2:

            teacherMob.Command('say You\'re ready! Riley\'s world is waiting for you.', extraDelay+1.0);
            teacherMob.Command('say I\'ll open a portal to Riley\'s neighborhood. Good luck out there!', extraDelay+2.0);

            exits = room.GetExits();
            if ( !exits.portal ) {
                teacherMob.Command('emote glows intensely, and a ' + UtilApplyColorPattern('swirling portal', 'rainbow') + ' appears!', extraDelay+3.0);
                room.AddTemporaryExit('swirling portal', ':rainbow', 0, 9000); // RoomId 0 is an alias for start room
            }

            teacherMob.Command('say Enter the portal by typing <ansi fg="command">swirling portal</ansi> (or just <ansi fg="command">portal</ansi>) when you\'re ready.', extraDelay+4.0);
            teacherMob.Command('say Remember -- every feeling matters. Even the scary ones. Especially the scary ones.', extraDelay+6.0);

            break;
        default:
            break;
    }

    return ignoreCommand;
}



function onEnter(user, room) {

    teacherMob = getTeacher(room);
    clearGroundItems(room);

    sendWorkingCommands(user);

    itm = CreateItem(capItemId);
    teacherMob.GiveItem(itm);

    teacherMob.Command('emote appears in a ' + UtilApplyColorPattern("flash of light!", "rainbow"));

    teacherMob.Command('say Congratulations! You made it through training! Joy would be so proud right now.', 1.0);
    teacherMob.Command('say Here -- take this. Every adventurer needs a lucky cap.', 2.0);
    teacherMob.Command('drop cap', 3.0);
    teacherMob.Command('emote gestures to the <ansi fg="item">baseball cap</ansi> on the ground.', 4.0);
    teacherMob.Command('say type <ansi fg="command">get cap</ansi> to pick up the <ansi fg="item">baseball cap</ansi>.', 5.0);

    return true;
}

function onExit(user , room) {
    destroyTeacher(room);
    commandNow = 0;
}

function onLoad(room) {
    commandNow = 0;
}

function getTeacher(room) {
    var mobActor = room.GetMob(teacherMobId, true);
    mobActor.SetCharacterName(teacherName);
    return mobActor;
}

function destroyTeacher(room) {
    var mobActor = room.GetMob(teacherMobId);
    if ( mobActor != null ) {
        mobActor.Command('suicide vanish');
    }
}

function sendWorkingCommands(user) {

    ac = [];
    unlockedCommands = teach_commands.slice(0, commandNow);

    for (var i in allowed_commands ) {
        ac.push(allowed_commands[i]);
    }

    for ( i in unlockedCommands ) {
        ac.push(unlockedCommands[i]);
    }

    user.SendText("");
    user.SendText("");
    user.SendText('    <ansi fg="red">NOTE:</ansi> Most commands have been <ansi fg="203">DISABLED</ansi> and <ansi fg="203">WILL NOT WORK</ansi> until you <ansi fg="51">COMPLETE THIS TUTORIAL</ansi>!');
    user.SendText("");
    user.SendText("");

}

function clearGroundItems(room) {

    allGroundItems = room.GetItems();
    for ( var i in allGroundItems ) {
        room.DestroyItem(allGroundItems[i]);
    }

}
