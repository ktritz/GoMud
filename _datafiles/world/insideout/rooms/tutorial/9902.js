
const allowed_commands = ["help", "broadcast", "look", "status", "inventory", "experience", "conditions"];
const teach_commands = ["equip pencil", "attack dummy", "west"];
const teacherMobId = 97;
const dummyMobId = 98;
const teacherName = "Orb of Courage";
const firstItemId = 10001;

var commandNow = 0; // Which command they are on



// Generic Command Handler
function onCommand(cmd, rest, user, room) {

    ignoreCommand = false;

    teacherMob = getTeacher(room);

    var extraDelay = 0;

    if ( cmd == "south" && !canGoSouth ) {
        teacherMob.Command("say Not so hasty! Lets finish up here before you leave the area.");
        ignoreCommand = true;
    }

    fullCommand = cmd;
    if ( rest.length > 0 ) {
        fullCommand = cmd + ' ' + rest;
    }

    if ( teach_commands[commandNow] == fullCommand ) {

        teacherMob.Command("say Good job!", 1.0);

        extraDelay = 1.0;

        if ( fullCommand == "equip pencil" ) {
            teacherMob.Command('say Check it out! If you type <ansi fg="command">status</ansi> you\'ll see the pencil is equipped!', 2.0);
            teacherMob.Command('say It\'s not much, but even Joy started with just a smile.', 3.0);
            extraDelay = 3.0;
        }

        commandNow++;

        if ( cmd == "attack" ) {
            return false;
        }

    } else {

        if ( allowed_commands.includes(cmd) || teach_commands.slice(0, commandNow).includes(cmd) ) {
            return false;
        }

        ignoreCommand = true;
    }

    switch (commandNow) {
        case 0:

            if ( !user.HasItemId(firstItemId) ) {
                itm = CreateItem(firstItemId);
                user.GiveItem(itm);
            }

            teacherMob.Command('say Go ahead and equip that pencil you\'ve got. Type <ansi fg="command">equip pencil</ansi>.', extraDelay+1.0);
            break;
        case 1:

            getDummy(room);

            teacherMob.Command('say You may have noticed the <ansi fg="mobname">training dummy</ansi> here. It\'s basically a thought that forgot how to think.', extraDelay+1.0);
            teacherMob.Command('say Go ahead and engage in combat by typing <ansi fg="command">attack dummy</ansi>.', extraDelay+2.0);
            break;
        case 2:
            break;
        default:
            break;
    }

    return ignoreCommand;
}



function onEnter(user, room) {
    room.SetLocked("north", true);

    teacherMob = getTeacher(room);
    getDummy(room);

    sendWorkingCommands(user);

    teacherMob.Command('emote appears in a ' + UtilApplyColorPattern("flash of light!", "glowing"));

    teacherMob.Command('say Alright, time to learn how to stand your ground! Even Anger would be proud.', 1.0);

    if ( !user.HasItemId(firstItemId) ) {
        itm = CreateItem(firstItemId);
        user.GiveItem(itm);
    }

    teacherMob.Command('say Go ahead and equip that pencil. Type <ansi fg="command">equip pencil</ansi>.', 2.0);

    return true;
}

function onExit(user , room) {
    destroyTeacher(room);
    destroyDummy(room);
    canGoSouth = false;
    commandNow = 0;
}

function onLoad(room) {
    canGoSouth = false;
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

function getDummy(room) {
    return room.GetMob(dummyMobId, true);
}

function destroyDummy(room) {
    var mobActor = room.GetMob(dummyMobId);
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
