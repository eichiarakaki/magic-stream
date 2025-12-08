import Button from "react-bootstrap/Button";
import Container from "react-bootstrap/Container";
import Nav from "react-bootstrap/Nav";
import { Navbar } from "react-bootstrap";
import { useNavigate, NavLink } from "react-router-dom";
import useAuth from "../hooks/useAuth.tsx";

const Header = ({ handleLogout }) => {
  const navigate = useNavigate();
  const { auth } = useAuth();

  return (
    <Navbar
      bg={"dark"}
      variant={"dark"}
      expand={"lg"}
      sticky={"top"}
      className={"shadow-sm"}
    >
      <Container>
        <Navbar.Brand>Magic Stream</Navbar.Brand>

        <Navbar.Toggle aria-controls="main-navbar-nav" />
        <Navbar.Collapse id="main-navbar-nav">
          <Nav className={"me-auto"}>
            <Nav.Link as={NavLink} to={"/"}>
              Home
            </Nav.Link>
            <Nav.Link as={NavLink} to={"/recommended-movies"}>
              Recommended
            </Nav.Link>
          </Nav>
          <Nav className={"ms-auto align-items-center"}>
            {auth ? (
              <>
                <span>
                  Hello, <strong>{auth.first_name}</strong>
                </span>
                <Button
                  variant={"outline-light"}
                  size={"sm"}
                  onClick={() => {
                    handleLogout();
                  }}
                >
                  LogOut
                </Button>
              </>
            ) : (
              <>
                <Button
                  variant={"outline-info"}
                  size={"sm"}
                  className={"me-2"}
                  onClick={() => navigate("/login")}
                >
                  Login
                </Button>
                <Button
                  variant={"outline-info"}
                  size={"sm"}
                  className={"me-2"}
                  onClick={() => navigate("/register")}
                >
                  Register
                </Button>
              </>
            )}
          </Nav>
        </Navbar.Collapse>
      </Container>
    </Navbar>
  );
};

export default Header;
