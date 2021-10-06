var yearsForm = {
    removeYear: function () {
        var year = event.target.parentNode;
        year.remove();
    },

    addYear: function () {
        var yearsParent = document.getElementById('years');
        var year = document.getElementById('add-year-input').value;
        var yearInputs = yearsParent.querySelectorAll('.year-input');
        for (var yearInput of yearInputs) {
            var yearI = yearInput.value;
            if (yearI === year) {
                return;
            }
        }
        var newYear = yearsForm.createYear(year, 'true');
        newYear.querySelector('.year-radio').focus();
    },

    createYear: function (yearNum, active) {
        var template = document.getElementById('year-template');
        var clone = document.importNode(template.content, true);
        var newYear = clone.querySelector('.form-group');
        newYear.id = 'year-' + yearNum;
        newYear.querySelector('.year-radio').id = 'year-' + yearNum + '-active';
        newYear.querySelector('.year-radio').value = yearNum;
        if (active == 'true') {
            newYear.querySelector('.year-radio').checked = true;
        }
        newYear.querySelector('.year-label').htmlFor = 'year-' + yearNum + '-active';
        newYear.querySelector('.year-label').innerText = yearNum;
        newYear.querySelector('.year-input').value = yearNum;
        var years = document.getElementById('year-form-items');
        years.appendChild(clone);
        return newYear;
    },

    initYears: function () {
        var years = document.getElementById('year-form-items').children;
        for (var year of years) {
            var value = year.querySelector('.value').innerText;
            var active = year.querySelector('.active').innerText;
            var newYear = yearsForm.createYear(value, active);
            year.replaceWith(newYear);
        }
    },

    initAddYearInput: function () {
        document.getElementById('add-year-input').value = new Date().getYear() + 1900;
    },
};

yearsForm.initYears();
yearsForm.initAddYearInput();