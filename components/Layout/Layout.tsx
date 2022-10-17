import React, { FC } from "react"
import Header from "../Header/Header"
import Footer from "../Footer/Footer"
import styles from './Layout.module.css'
import HomeSlider from "../HomeSlider/HomeSlider";

type Props = {
  children?: React.ReactNode,
};

const Layout: FC<Props> = ({ children }) => {
  return (
    <main className={styles.layout}>
      <Header />
      {children}
      <Footer />
    </main>
  )
}

export default Layout