var playerSearch = {
    showModal: function (show) {
        var playerSearchModal = document.getElementById('player-search-modal');
        playerSearchModal.classList.toggle('d-none', !show);
        var playerSearchModalOpenButton = document.getElementById('player-search-modal-open')
        playerSearchModalOpenButton.classList.toggle('d-none', show);
        if (show) {
            playerSearch.initActivePlayersCB();
            var searchInput = document.getElementById('player-search');
            searchInput.focus();
        } else {
            playerSearchModal.focus();
        }
    },

    initActivePlayersCB: function() {
        var playerType = document.getElementById('select-player-type').value;
        var isMlbPlayerType = [2, 3].includes(parseInt(playerType)); // PlayerTypeHitter, PlayerTypePitcher
        var activePlayersOnlyGroup = document.getElementById('apo-group');
        activePlayersOnlyGroup.classList.toggle('d-none', !isMlbPlayerType);
    },

    add: function () {
        var playerSearchResults = document.getElementById('player-search-results');
        playerSearchResults = playerSearchResults.getElementsByClassName('form-group');
        for (var i = 0; i < playerSearchResults.length; i++) {
            var psr = playerSearchResults[i];
            if (psr.querySelector('.psr-radio').checked) {
                var sourceID = psr.querySelector('.psr-source-id').value;
                var playerName = psr.querySelector('.psr-player-name').value;
                var newPlayer = playersForm.add(playerName, sourceID);
                newPlayer.focus();
                playerSearch.showModal(false);
                playerSearch.clear();
                return;
            }
        }
    },

    clear: function () {
        var resultsDiv = document.getElementById('player-search-results-output');
        resultsDiv.innerHTML = '';
        var searchInput = document.getElementById('player-search');
        searchInput.value = '';
    },

    submit: function (event) {
        event.preventDefault();
        playerSearch.setNewPlayerSelected(false);
        var playerType = document.getElementById('select-player-type').value;
        var formData = new FormData(event.target);
        formData.append('pt', playerType);
        var url = window.location.pathname + '/search?' + new URLSearchParams(formData);
        fetch(url, {
            method: 'GET',
        }).then(async res => {
            if (res.status == 200) {
                return res.json();
            } else {
                var message = await res.text();
                return Promise.reject(message);
            }
        }).then(playerSearchResults => {
            if (playerSearchResults != null && playerSearchResults.length > 0) {
                playerSearch.success(playerSearchResults);
                return Promise.resolve();
            } else {
                return Promise.reject('No results');
            }
        }).catch(err => {
            var resultsDiv = document.getElementById('player-search-results-output');
            resultsDiv.innerHTML = err;
        });
    },

    success: function (playerSearchResults) {
        var template = document.getElementById('player-search-results-template');
        var playerSearchResultsDiv = document.importNode(template.content, true);
        var playerSearchResultsFieldSet = playerSearchResultsDiv.getElementById('player-search-results');
        for (var i = 0; i < playerSearchResults.length; i++) {
            var playerSearchResult = playerSearchResults[i];
            var template2 = document.getElementById('player-search-result-template');
            var playerSearchResultDiv = document.importNode(template2.content, true);
            var psr = playerSearchResultDiv.querySelector('.form-group');
            psr.querySelector('.psr-radio').id = 'psr-' + playerSearchResult.SourceID;
            psr.querySelector('.psr-label').htmlFor = 'psr-' + playerSearchResult.SourceID;
            psr.querySelector('.psr-label-name').innerText = playerSearchResult.Name;
            psr.querySelector('.psr-label-details').innerText = playerSearchResult.Details;
            psr.querySelector('.psr-source-id').value = playerSearchResult.SourceID;
            psr.querySelector('.psr-player-name').value = playerSearchResult.Name;
            playerSearchResultsFieldSet.appendChild(psr);
            if (i == 0 && playerSearchResults.length == 1) {
                psr.querySelector('.psr-radio').checked = true;
            }
        }
        var resultsDiv = document.getElementById('player-search-results-output');
        resultsDiv.innerHTML = '';
        resultsDiv.appendChild(playerSearchResultsFieldSet);
        playerSearch.setNewPlayerSelected(playerSearchResults.length == 1);
    },

    setNewPlayerSelected: function (isSelected) {
        var addPlayerButton = document.getElementById('modal-add-player-button');
        addPlayerButton.disabled = !isSelected;
    },
};
