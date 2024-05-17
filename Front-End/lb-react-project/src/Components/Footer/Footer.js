import './Footer.css';

function Footer() {
    return (
        <section id="footer">
            <footer className="footer">
                <div className="footerContent">
                    © {new Date().getFullYear()} My Company - All Rights Reserved
                </div>
            </footer>
        </section>
    );
}

export default Footer;