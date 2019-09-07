var adminFormItem = {
    moveUp: function (event) {
        var item1 = event.target.parentNode;
        var item2 = event.target.parentNode.previousElementSibling;
        if (item2 != null) {
            var items = item1.parentNode;
            items.insertBefore(item1, item2);
            adminFormItem.changeDisplayOrder(item1, -1);
            adminFormItem.changeDisplayOrder(item2, +1);
        }
    },

    moveDown: function (event) {
        var item1 = event.target.parentNode;
        var item2 = event.target.parentNode.nextElementSibling;
        if (item2 != null) {
            var items = item1.parentNode;
            items.insertBefore(item2, item1);
            adminFormItem.changeDisplayOrder(item1, +1);
            adminFormItem.changeDisplayOrder(item2, -1);
        }
    },

    changeDisplayOrder: function (item, delta) {
        var itemDisplayOrderElement = item.querySelector('.admin-form-item-display-order');
        var itemDisplayOrder = itemDisplayOrderElement.value;
        itemDisplayOrderElement.value = parseInt(itemDisplayOrder) + delta;
    },

    remove: function (event) {
        var item = event.target.parentNode;
        var nextItem = document.getElementById(item.id).nextElementSibling;
        while (nextItem != null) {
            adminFormItem.changeDisplayOrder(nextItem, -1);
            nextItem = nextItem.nextElementSibling;
        }
        item.remove();
        return item;
    },
};