let cancelModal
document.addEventListener("DOMContentLoaded", ()=>{
    cancelModal = document.getElementById("cancel-modal")
    document.body.addEventListener("htmx:afterSettle", ()=>{
        cancelModal = document.getElementById("cancel-modal")
    })
})

const onOverlayClick = (ev)=>{
    if (ev.target.id === "cancel-modal"){
        cancelModal.style.display = "none"
    }
}

const onOpenCancelModal = (ele)=>{
    cancelModal.style.display = "flex"
    cancelModal.innerHTML = `
        <div class="modal">
            <div class="modal-header">
                <h3 class="modal-title">Annuler le rendez-vous</h3>
            </div>
            <div class="modal-body">
                <p class="modal-text">Êtes-vous sûr de vouloir annuler ce rendez-vous ? Cette action est irréversible.</p>
            </div>
            <div class="modal-footer">
                <button type="button" class="btn btn-outline" id="cancel-modal-close" onclick="cancelModal.style.display = 'none'">Retour</button>
                <button type="button" class="btn btn-danger" id="confirm-cancel" hx-delete="/rendez-vous/${ele.dataset.id}">Confirmer l'annulation</button>
            </div>
        </div>
        `
    htmx.process(document.body)
}
