var playersForm = {
    add: function (playerName, sourceID) {
        var playerType = document.getElementById('select-player-type').value;
        var friendID = document.getElementById('select-friend').value;
        var maxID = 0;
        var maxDisplayOrder = 0;
        var players = document.getElementById('player-form-items');
        var playerElements = players.getElementsByClassName('form-group');
        for (var player of playerElements) {
            var ID = parseInt(player.querySelector('.player-id').value);
            if (ID > maxID) {
                maxID = ID;
            }
            if (playerType === player.querySelector('.player-player-type').value &&
                friendID === player.querySelector('.player-friend-id').value) {
                var playerDisplayOrder = parseInt(player.querySelector('.player-display-order').value);
                if (playerDisplayOrder > maxDisplayOrder) {
                    maxDisplayOrder = playerDisplayOrder;
                }
            }
        }
        return playersForm.create(maxID + 1, playerName, sourceID, maxDisplayOrder + 1, playerType, friendID);
    },

    create: function (id, playerName, sourceID, displayOrder, playerType, friendID) {
        var template = document.getElementById('player-template');
        var clone = document.importNode(template.content, true);
        var player = clone.querySelector('.form-group');
        player.id = 'player-' + id;
        player.querySelector('.player-name-label').innerText = playerName;
        player.querySelector('.player-source-id').name = 'player-' + id + '-source-id';
        player.querySelector('.player-source-id').value = sourceID;
        player.querySelector('.player-display-order').name = 'player-' + id + '-display-order';
        player.querySelector('.player-display-order').value = displayOrder;
        player.querySelector('.player-player-type').name = 'player-' + id + '-player-type';
        player.querySelector('.player-player-type').value = playerType;
        player.querySelector('.player-friend-id').name = 'player-' + id + '-friend-id';
        player.querySelector('.player-friend-id').value = friendID;
        player.querySelector('.player-id').value = id;
        var scoreCategories = document.getElementById('players');
        var scoreCategory = scoreCategories.querySelector('.player-type-' + playerType);
        var friendScore = scoreCategory.querySelector('.friend-id-' + friendID);
        friendScore.appendChild(clone);
        return player;
    },

    refresh: function () {
        var playerType = document.getElementById('select-player-type').value;
        var friendID = document.getElementById('select-friend').value;
        var players = document.getElementById('player-form-items');
        var playerElements = players.getElementsByClassName('form-group');
        for (var player of playerElements) {
            var currentPlayerType = player.querySelector('.player-player-type').value;
            var currentFriendID = player.querySelector('.player-friend-id').value;
            if (playerType === currentPlayerType && friendID === currentFriendID) {
                player.classList.remove('d-none');
            } else {
                player.classList.add('d-none');
            }
        }
        playerSearch.clear();
    },

    init: function () {
        var friendSelect = document.getElementById('select-friend');
        var playerTypeSelect = document.getElementById('player-template');
        if (friendSelect == null) {
            var message = playerTypeSelect == null
                ? 'Requires Year'
                : 'Requires Friend';
            adminFormItem.disableButtons(['player-search-modal-open', 'players-form-submit-button'], message)
            return;
        }
        var playerTypes = document.getElementById('player-form-items').children;
        for (var playerType of playerTypes) {
            var friendScores = playerType.children;
            for (var friendScore of friendScores) {
                var playerScores = friendScore.children;
                for (var playerScore of playerScores) {
                    var id = playerScore.querySelector('.id').innerText;
                    var playerName = playerScore.querySelector('.player-name').innerText;
                    var sourceID = playerScore.querySelector('.source-id').innerText;
                    var displayOrder = playerScore.querySelector('.display-order').innerText;
                    var pt = playerScore.querySelector('.player-type').innerText;
                    var friendID = playerScore.querySelector('.friend-id').innerText;
                    var newPlayer = playersForm.create(id, playerName, sourceID, displayOrder, pt, friendID);
                    playerScore.replaceWith(newPlayer);
                }
            }
        }
        playersForm.refresh();
    },
}

playersForm.init();
