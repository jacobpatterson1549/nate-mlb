var friendsForm = {
    add: function () {
        var maxFriendId = 0;
        var friendsParent = document.getElementById('friends');
        var friendIdInputs = friendsParent.querySelectorAll('.friend-id');
        for (var friendIdInput of friendIdInputs) {
            var friendId = parseInt(friendIdInput.value);
            if (friendId > maxFriendId) {
                maxFriendId = friendId;
            }
        }
        var newFriend = friendsForm.create(maxFriendId + 1, '', friendIdInputs.length, '[NEW]');
        newFriend.querySelector('.friend-name-input').focus();
    },

    create: function (id, name, displayOrder, nameLabel) {
        var template = document.getElementById('friend-template');
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
        var friends = document.getElementById('friend-form-items');
        friends.appendChild(clone);
        return friend;
    },

    init: function () {
        if (document.getElementById('friend-template') == null) {
            adminFormItem.disableButtons(['add-friend-button', 'friends-form-submit-button'], 'Requires Year');
            return;
        }
        var friends = document.getElementById('friend-form-items').children;
        for (var friend of friends) {
            var id = friend.querySelector('.id').innerText;
            var name = friend.querySelector('.name').innerText;
            var displayOrder = friend.querySelector('.displayOrder').innerText;
            var newFriend = friendsForm.create(id, name, displayOrder, name);
            friend.replaceWith(newFriend);
        }
    },
};

friendsForm.init();