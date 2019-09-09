var playersForm = {
    add: function (playerName, sourceID) {
        var playerType = document.getElementById('select-player-type').value;
        var friendID = document.getElementById('select-friend').value;
        var maxID = 0;
        var maxDisplayOrder = 0;
        var players = document.getElementById('player-form-items');
        var playerElements = players.getElementsByClassName('form-group');
        for (var i = 0; i < playerElements.length; i++) {
            var player = playerElements[i];
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
        var template = document.getElementById("player-template");
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
        var scoreCategories = document.getElementById("players");
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
        for (var i = 0; i < playerElements.length; i++) {
            var player = playerElements[i];
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
        if (document.getElementById('select-friend') == null) {
            var message = document.getElementById('player-template') == null
                ? "Requires Year"
                : "Requires Friend";
            adminFormItem.disableButtons(['openPlayerSearchModal', 'players-form-submit-button'], message)
            return;
        }
        var playerTypes = document.getElementById("player-form-items").children;
        for (var i = 0; i < playerTypes.length; i++) {
            var friendScores = playerTypes[i].children;
            for (var j = 0; j < friendScores.length; j++) {
                var playerScores = friendScores[j].children;
                for (var k = 0; k < playerScores.length; k++) {
                    var id = playerScores[k].querySelector(".id").innerText;
                    var playerName = playerScores[k].querySelector(".playerName").innerText;
                    var sourceID = playerScores[k].querySelector(".sourceID").innerText;
                    var displayOrder = playerScores[k].querySelector(".displayOrder").innerText;
                    var playerType = playerScores[k].querySelector(".playerType").innerText;
                    var friendID = playerScores[k].querySelector(".friendID").innerText;
                    var newPlayer = playersForm.create(id, playerName, sourceID, displayOrder, playerType, friendID);
                    playerScores[k].replaceWith(newPlayer);
                }
            }
        }
        playersForm.refresh();
    },
}

playersForm.init();
