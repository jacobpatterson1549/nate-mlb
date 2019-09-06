function moveFriendUp() {
    var friend1 = event.target.parentNode;
    var friend2 = event.target.parentNode.previousElementSibling;
    if (friend2 != null) {
        var friends = friend1.parentNode;
        friends.insertBefore(friend1, friend2);
        changeFriendDisplayOrder(friend1.id, -1);
        changeFriendDisplayOrder(friend2.id, +1);
    }
}

function moveFriendDown() {
    var friend1 = event.target.parentNode;
    var friend2 = event.target.parentNode.nextElementSibling;
    if (friend2 != null) {
        var friends = friend1.parentNode;
        friends.insertBefore(friend2, friend1);
        changeFriendDisplayOrder(friend2.id, -1);
        changeFriendDisplayOrder(friend1.id, +1);
    }
}

function removeFriend() {
    var removeFriend = event.target.parentNode;
    for (var friend = document.getElementById(removeFriend.id).nextElementSibling; friend != null; friend = document
        .getElementById(friend.id).nextElementSibling) {
        changeFriendDisplayOrder(friend.id, -1);
    }
    removeFriend.remove();
}

function addFriend() {
    var maxFriendId = 0;
    var friendsParent = document.getElementById('friends');
    var friendIdInputs = friendsParent.querySelectorAll('.friend-id');
    for (var i = 0; i < friendIdInputs.length; i++) {
        var friendId = parseInt(friendIdInputs[i].value);
        if (friendId > maxFriendId) {
            maxFriendId = friendId;
        }
    }
    var newFriend = createFriend(maxFriendId + 1, '', friendIdInputs.length, '[NEW]');
    newFriend.querySelector('.friend-name-input').focus();
}

function createFriend(id, name, displayOrder, nameLabel) {
    var template = document.getElementById("friend-template");
    var clone = document.importNode(template.content, true);
    var friend = clone.querySelector('.form-group');
    friend.id = 'friend-' + id;
    friend.querySelector('.friend-name-label').htmlFor = 'friend-' + id + '-name';
    friend.querySelector('.friend-name-label').innerText = nameLabel;
    friend.querySelector('.friend-name-input').id = 'friend-' + id + '-name';
    friend.querySelector('.friend-name-input').name = 'friend-' + id + '-name';
    friend.querySelector('.friend-name-input').value = name;
    friend.querySelector('.friend-display-order').name = 'friend-' + id + '-display-order';
    friend.querySelector('.friend-display-order').value = displayOrder;
    friend.querySelector('.friend-id').value = id;
    var friends = document.getElementById("friends");
    friends.appendChild(clone);
    return friend;
}

function changeFriendDisplayOrder(friendID, delta) {
    var friend = document.getElementById(friendID);
    var friendDisplayOrderElement = friend.querySelector('.friend-display-order');
    var friendDisplayOrder = friendDisplayOrderElement.value;
    friendDisplayOrderElement.value = parseInt(friendDisplayOrder) + delta;
}

function initFriends() {
    var friends = document.getElementById("friends").children;
    for (var i = 0; i < friends.length; i++) {
        var id = friends[i].querySelector(".id").innerText;
        var name = friends[i].querySelector(".name").innerText;
        var displayOrder = friends[i].querySelector(".displayOrder").innerText;
        var newFriend = createFriend(id, name, displayOrder, name);
        friends[i].replaceWith(newFriend);
    }
}

initFriends();