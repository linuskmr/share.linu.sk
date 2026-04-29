/**
 * Port of the Go RoundNicely logic.
 */
function roundNicely(n) {
    let precision;
    if (n >= 10) {
        precision = 0;
    } else if (n >= 1) {
        precision = 1;
    } else {
        precision = 2;
    }

    const ratio = Math.pow(10, precision);
    const rounded = Math.round(n * ratio) / ratio;
    return rounded.toFixed(precision).replace('.', ',');
}

/**
 * Main function to scale ingredients and inject the interactive input.
 */
function setupInteractiveRecipe() {
    const mainElement = document.querySelector('main');
    if (!mainElement) return;

    // 1. Setup URL and Portions
    const url = new URL(window.location.href);
    const urlQuantity = parseInt(url.searchParams.get('quantity'));
    
    const portionRegex = /(\d+)\s+Portionen/i;
    const portionMatch = mainElement.innerText.match(portionRegex);
    if (!portionMatch) return;

    const basePortions = parseInt(portionMatch[1]);
    // Use URL quantity if present, otherwise fallback to the document's base
    const currentPortions = urlQuantity || basePortions;
    const scalingRatio = currentPortions / basePortions;

    // 2. Replace static text with an <input> field
    // We use innerHTML carefully here to inject the input element
    mainElement.innerHTML = mainElement.innerHTML.replace(portionRegex, 
        `<input type="number" id="portion-scaler" value="${currentPortions}" min="1" style="width: 50px;"> Portionen`
    );

    // Add event listener to the new input
    document.getElementById('portion-scaler').addEventListener('change', (e) => {
        const newQty = e.target.value;
        url.searchParams.set('quantity', newQty);
        window.location.href = url.toString(); // Reloads with new quantity
    });

    // 3. If we are scaled (ratio != 1), recalculate ingredient quantities
    if (scalingRatio !== 1) {
        const ingredientRegex = /(\d+(?:,\d+)?)\s?(ml|g|kg|l|x|TL|EL)\s+([a-zA-ZäöüÄÖÜß]+)/g;
        const walker = document.createTreeWalker(mainElement, NodeFilter.SHOW_TEXT, null, false);
        let textNode;

        while (textNode = walker.nextNode()) {
            // Skip the text inside our new input if the walker hits it
            if (textNode.parentElement.id === 'portion-scaler') continue;

            textNode.nodeValue = textNode.nodeValue.replace(ingredientRegex, (match, amountStr, unit, name) => {
                const originalAmount = parseFloat(amountStr.replace(',', '.'));
                const scaledAmount = originalAmount * scalingRatio;
                return `${roundNicely(scaledAmount)} ${unit} ${name}`;
            });
        }
    }
}

document.addEventListener('DOMContentLoaded', setupInteractiveRecipe);